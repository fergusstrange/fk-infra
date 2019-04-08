package templates

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/infinityworks/fk-infra/aws"
	"github.com/infinityworks/fk-infra/model"
	"github.com/infinityworks/fk-infra/util"
	"log"
	"text/template"
)

const elasticSearchTemplate = `
terraform {
  backend "s3" {
    bucket = "{{$.ConfigBucket}}"
    key    = "terraform/elasticsearch/terraform.tfstate"
    region = "{{$.Region}}"
  }

  required_version = ">= 0.9.3"
}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

{{range .Clusters}}
resource "aws_security_group" "{{$.EnvironmentName}}-{{.Name}}-elasticsearch" {
  name = "{{$.EnvironmentName}}-{{.Name}}-elasticsearch"
  description = "Managed by Terraform"
  vpc_id = "${aws_vpc.{{$.EnvironmentName}}.id}"

  ingress {
    from_port = 443
    to_port = 443
    protocol = "tcp"

    security_groups = [
      "${aws_security_group.k8s-masters-{{$.EnvironmentName}}.id}",
      "${aws_security_group.k8s-nodes-{{$.EnvironmentName}}.id}"
    ]
  }
}

resource "aws_elasticsearch_domain" "{{$.EnvironmentName}}-{{.Name}}" {
  domain_name = "{{$.EnvironmentName}}-{{.Name}}"
  elasticsearch_version = "6.4"

  cluster_config {
    instance_type = "m3.medium.elasticsearch"
    dedicated_master_enabled = false
    zone_awareness_enabled = true
    instance_count = 4
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 100
    volume_type = "gp2"
  }

  vpc_options {
    subnet_ids = [
      "${aws_subnet.{{$.Region}}a-{{$.EnvironmentName}}.id}",
      "${aws_subnet.{{$.Region}}b-{{$.EnvironmentName}}.id}"
    ]

    security_group_ids = [
      "${aws_security_group.{{$.EnvironmentName}}-{{.Name}}-elasticsearch.id}"]
  }

  advanced_options {
    "rest.action.multi.allow_explicit_index" = "true"
  }

  access_policies = <<CONFIG
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "es:*",
            "Principal": "*",
            "Effect": "Allow",
            "Resource": "arn:aws:es:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:domain/{{.Name}}/*"
        }
    ]
}
CONFIG

  snapshot_options {
    automated_snapshot_start_hour = 3
  }

  tags {
    Name = "{{.Name}}"
    Domain = "{{.Name}}"
  }
}

output "elasticsearch_output_{{.Name}}" {
  value = "{\"name\":\"{{.Name}}\",\"endpoint\":\"${aws_elasticsearch_domain.{{$.EnvironmentName}}-{{.Name}}.endpoint}\",\"arn\":\"${aws_elasticsearch_domain.{{$.EnvironmentName}}-{{.Name}}.arn}\"}"
}
{{end}}
`

func RenderElasticSearch(config *model.Config) {
	createAwsElasticSearchServiceRole(config.Spec.Region)
	terraformTemplate := parseElasticSearchTemplate(
		config.Spec.EnvironmentName,
		config.Spec.Region,
		config.Spec.ConfigBucket,
		config.Spec.ElasticSearch)
	util.WriteFile("./elasticsearch.tf", terraformTemplate)
}

func createAwsElasticSearchServiceRole(region string) {
	_, err := iam.New(aws.NewSession(region)).CreateServiceLinkedRole(&iam.CreateServiceLinkedRoleInput{
		AWSServiceName: util.String("es.amazonaws.com"),
	})
	if errWithCode, ok := err.(awserr.Error); ok && iam.ErrCodeInvalidInputException == errWithCode.Code() {
		log.Println("ElasticSearch service role already exists")
	} else {
		util.CheckError(errWithCode)
	}
}

func parseElasticSearchTemplate(environmentName, region, configBucket string, elasticSearchSpec []model.ElasticSearch) []byte {
	var buf bytes.Buffer
	tmpl, err := template.New("elasticSearchTemplate").Parse(elasticSearchTemplate)
	util.CheckError(err)
	err = tmpl.Execute(&buf, ElasticSearchTemplate{
		EnvironmentName: environmentName,
		Region:          region,
		ConfigBucket:    configBucket,
		Clusters:        clusterTemplates(elasticSearchSpec),
	})
	util.CheckError(err)
	return buf.Bytes()
}

func clusterTemplates(elasticSearchClusters []model.ElasticSearch) []ElasticSearchClusterTemplate {
	var elasticSearchClusterTemplates []ElasticSearchClusterTemplate
	for _, cluster := range elasticSearchClusters {
		elasticSearchClusterTemplates = append(elasticSearchClusterTemplates, ElasticSearchClusterTemplate{
			Name: cluster.Name,
		})
	}
	return elasticSearchClusterTemplates
}

type ElasticSearchClusterTemplate struct {
	Name string
}

type ElasticSearchTemplate struct {
	EnvironmentName string
	Region          string
	ConfigBucket    string
	Clusters        []ElasticSearchClusterTemplate
}
