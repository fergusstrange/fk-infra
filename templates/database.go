package templates

import (
	"bytes"
	"github.com/infinityworks/fk-infra/model"
	"github.com/infinityworks/fk-infra/terraform"
	"github.com/infinityworks/fk-infra/util"
	"text/template"
)

const databaseTemplate = `
terraform {
  backend "s3" {
    bucket = "{{$.ConfigBucket}}"
    key    = "terraform/database/terraform.tfstate"
    region = "{{$.Region}}"
  }

  required_version = ">= 0.9.3"
}

{{range .Databases}}
output "database_output_{{.Name}}" {
  value = "{\"name\":\"{{.Name}}\",\"endpoint\":\"${aws_db_instance.{{$.EnvironmentName}}-{{.Name}}.endpoint}\",\"password\":\"${aws_db_instance.{{$.EnvironmentName}}-{{.Name}}.password}\"}"
}

resource "aws_security_group" "{{$.EnvironmentName}}-{{.Name}}-database-access" {
  vpc_id = "${aws_vpc.{{$.EnvironmentName}}.id}"
  name = "{{$.EnvironmentName}}-{{.Name}}-database-access"
  description = "Allow access to database"
  ingress {
      from_port = 3306
      to_port = 3306
      protocol = "tcp"
      security_groups = ["${aws_security_group.k8s-nodes-{{$.EnvironmentName}}.id}"]
  }
  egress {
      from_port = 0
      to_port = 0
      protocol = "-1"
      cidr_blocks = ["0.0.0.0/0"]
      self = true
  }
  tags {
    Name = "{{.Name}}-database-access"
  }
}

resource "aws_db_subnet_group" "{{$.EnvironmentName}}-{{.Name}}" {
    name = "{{.Name}}-subnet"
    description = "RDS subnet group"
    subnet_ids = ["${aws_subnet.{{$.Region}}a-{{$.EnvironmentName}}.id}","${aws_subnet.{{$.Region}}b-{{$.EnvironmentName}}.id}"]
}

resource "aws_db_parameter_group" "{{$.EnvironmentName}}-{{.Name}}" {
    name = "{{$.EnvironmentName}}-{{.Name}}"
    family = "mysql8.0"
    description = "Mysql parameter group"

    parameter {
      name = "slow_query_log"
      value = 1
   }

}

resource "aws_db_instance" "{{$.EnvironmentName}}-{{.Name}}" {
  allocated_storage    = 120
  engine               = "mysql"
  engine_version       = "8.0"
  instance_class       = "db.t3.small"
  identifier           = "{{$.EnvironmentName}}-{{.Name}}"
  name                 = "{{.Name}}"
  username             = "{{.Name}}"
  password             = "{{.Password}}"
  db_subnet_group_name = "${aws_db_subnet_group.{{$.EnvironmentName}}-{{.Name}}.name}"
  parameter_group_name = "${aws_db_parameter_group.{{$.EnvironmentName}}-{{.Name}}.name}"
  multi_az             = "true"
  vpc_security_group_ids = ["${aws_security_group.{{$.EnvironmentName}}-{{.Name}}-database-access.id}"]
  storage_type         = "gp2"
  backup_retention_period = 30
  final_snapshot_identifier = "{{$.EnvironmentName}}-{{.Name}}-final-snapshot"
  tags {
      Name = "{{$.EnvironmentName}}-{{.Name}}"
  }
}

{{end}}
`

func RenderDatabases(config *model.Config) {
	var databaseTemplates []DatabaseTemplate
	databaseOutputs := terraform.FetchTerraformOutputs().DatabaseConfig()
	for _, database := range config.Spec.Databases {
		databaseTemplates = append(databaseTemplates, DatabaseTemplate{
			Name:     database.Name,
			Password: fetchOrGeneratePassword(databaseOutputs, database.Name),
		})
	}

	databaseTemplate := parseDatabasesTemplate(DatabasesTemplate{
		ConfigBucket:    config.Spec.ConfigBucket,
		Region:          config.Spec.Region,
		EnvironmentName: config.Spec.EnvironmentName,
		Databases:       databaseTemplates,
	})

	util.WriteFile("./databases.tf", databaseTemplate)
}

func parseDatabasesTemplate(databasesTemplate DatabasesTemplate) []byte {
	var buf bytes.Buffer
	tmpl, err := template.New("databaseTemplate").Parse(databaseTemplate)
	util.CheckError(err)
	util.CheckError(tmpl.Execute(&buf, databasesTemplate))
	return buf.Bytes()
}

func fetchOrGeneratePassword(databaseOutputs []terraform.DatabaseOutput, databaseName string) string {
	for _, databaseOutput := range databaseOutputs {
		if databaseOutput.Name == databaseName {
			return databaseOutput.Password
		}
	}
	return util.RandomAlphaNumeric(32)
}

type DatabasesTemplate struct {
	Region          string
	ConfigBucket    string
	EnvironmentName string
	Databases       []DatabaseTemplate
}

type DatabaseTemplate struct {
	Name     string
	Password string
}
