package templates

import (
	"bytes"
	"github.com/infinityworks/fk-infra/model"
	"github.com/infinityworks/fk-infra/util"
	"html/template"
)

const networkTemplate = `
provider "aws" {
  region = "{{.Region}}"
}

terraform {
  backend "s3" {
    bucket = "{{.ConfigBucket}}"
    key    = "terraform/network/terraform.tfstate"
    region = "{{.Region}}"
  }

  required_version = ">= 0.9.3"
}

locals = {
  cluster_name                      = "{{.EnvironmentName}}"
  node_subnet_ids                   = ["${aws_subnet.{{.Region}}a-{{.EnvironmentName}}.id}", "${aws_subnet.{{.Region}}b-{{.EnvironmentName}}.id}"]
  region                            = "{{.Region}}"
  route_table_private-{{.Region}}a_id = "${aws_route_table.private-{{.Region}}a-{{.EnvironmentName}}.id}"
  route_table_private-{{.Region}}b_id = "${aws_route_table.private-{{.Region}}b-{{.EnvironmentName}}.id}"
  route_table_public_id             = "${aws_route_table.{{.EnvironmentName}}.id}"
  subnet_{{.Region}}a_id              = "${aws_subnet.{{.Region}}a-{{.EnvironmentName}}.id}"
  subnet_{{.Region}}b_id              = "${aws_subnet.{{.Region}}b-{{.EnvironmentName}}.id}"
  subnet_utility-{{.Region}}a_id      = "${aws_subnet.utility-{{.Region}}a-{{.EnvironmentName}}.id}"
  subnet_utility-{{.Region}}b_id      = "${aws_subnet.utility-{{.Region}}b-{{.EnvironmentName}}.id}"
  vpc_cidr_block                    = "${aws_vpc.{{.EnvironmentName}}.cidr_block}"
  vpc_id                            = "${aws_vpc.{{.EnvironmentName}}.id}"
}

output "cluster_name" {
  value = "{{.EnvironmentName}}"
}

output "node_subnet_ids" {
  value = ["${aws_subnet.{{.Region}}a-{{.EnvironmentName}}.id}", "${aws_subnet.{{.Region}}b-{{.EnvironmentName}}.id}"]
}

output "region" {
  value = "{{.Region}}"
}

output "route_table_private-{{.Region}}a_id" {
  value = "${aws_route_table.private-{{.Region}}a-{{.EnvironmentName}}.id}"
}

output "route_table_private-{{.Region}}b_id" {
  value = "${aws_route_table.private-{{.Region}}b-{{.EnvironmentName}}.id}"
}

output "route_table_public_id" {
  value = "${aws_route_table.{{.EnvironmentName}}.id}"
}

output "subnet_a" {
  value = "${aws_subnet.{{.Region}}a-{{.EnvironmentName}}.id}"
}

output "subnet_b" {
  value = "${aws_subnet.{{.Region}}b-{{.EnvironmentName}}.id}"
}

output "subnet_utility-a" {
  value = "${aws_subnet.utility-{{.Region}}a-{{.EnvironmentName}}.id}"
}

output "subnet_utility-b" {
  value = "${aws_subnet.utility-{{.Region}}b-{{.EnvironmentName}}.id}"
}

output "vpc_cidr_block" {
  value = "${aws_vpc.{{.EnvironmentName}}.cidr_block}"
}

output "vpc_id" {
  value = "${aws_vpc.{{.EnvironmentName}}.id}"
}

output "worker_security_group_id" {
  value = "${aws_security_group.k8s-nodes-{{.EnvironmentName}}.id}"
}

output "master_security_group_id" {
  value = "${aws_security_group.k8s-masters-{{.EnvironmentName}}.id}"
}

resource "aws_eip" "{{.Region}}a-{{.EnvironmentName}}" {
  vpc = true

  tags = {
    Name                                        = "{{.Region}}a.{{.EnvironmentName}}"
  }
}

resource "aws_eip" "{{.Region}}b-{{.EnvironmentName}}" {
  vpc = true

  tags = {
    Name                                        = "{{.Region}}b.{{.EnvironmentName}}"
  }
}

resource "aws_security_group" "k8s-masters-{{.EnvironmentName}}" {
  name        = "masters.lol.k8s.local"
  vpc_id      = "${aws_vpc.{{.EnvironmentName}}.id}"
  description = "Security group for masters"
}

resource "aws_security_group" "k8s-nodes-{{.EnvironmentName}}" {
  name        = "nodes.lol.k8s.local"
  vpc_id      = "${aws_vpc.{{.EnvironmentName}}.id}"
  description = "Security group for nodes"
}

resource "aws_internet_gateway" "{{.EnvironmentName}}" {
  vpc_id = "${aws_vpc.{{.EnvironmentName}}.id}"

  tags = {
    Name                                        = "{{.EnvironmentName}}"
  }
}

resource "aws_nat_gateway" "{{.Region}}a-{{.EnvironmentName}}" {
  allocation_id = "${aws_eip.{{.Region}}a-{{.EnvironmentName}}.id}"
  subnet_id     = "${aws_subnet.utility-{{.Region}}a-{{.EnvironmentName}}.id}"

  tags = {
    Name                                        = "{{.Region}}a.{{.EnvironmentName}}"
  }
}

resource "aws_nat_gateway" "{{.Region}}b-{{.EnvironmentName}}" {
  allocation_id = "${aws_eip.{{.Region}}b-{{.EnvironmentName}}.id}"
  subnet_id     = "${aws_subnet.utility-{{.Region}}b-{{.EnvironmentName}}.id}"

  tags = {
    Name                                        = "{{.Region}}b.{{.EnvironmentName}}"
  }
}

resource "aws_route" "0-0-0-0--0" {
  route_table_id         = "${aws_route_table.{{.EnvironmentName}}.id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.{{.EnvironmentName}}.id}"
}

resource "aws_route" "private-{{.Region}}a-0-0-0-0--0" {
  route_table_id         = "${aws_route_table.private-{{.Region}}a-{{.EnvironmentName}}.id}"
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = "${aws_nat_gateway.{{.Region}}a-{{.EnvironmentName}}.id}"
}

resource "aws_route" "private-{{.Region}}b-0-0-0-0--0" {
  route_table_id         = "${aws_route_table.private-{{.Region}}b-{{.EnvironmentName}}.id}"
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = "${aws_nat_gateway.{{.Region}}b-{{.EnvironmentName}}.id}"
}

resource "aws_route_table" "private-{{.Region}}a-{{.EnvironmentName}}" {
  vpc_id = "${aws_vpc.{{.EnvironmentName}}.id}"

  tags = {
    Name                                        = "private-{{.Region}}a.{{.EnvironmentName}}"
  }
}

resource "aws_route_table" "private-{{.Region}}b-{{.EnvironmentName}}" {
  vpc_id = "${aws_vpc.{{.EnvironmentName}}.id}"

  tags = {
    Name                                        = "private-{{.Region}}b.{{.EnvironmentName}}"
  }
}

resource "aws_route_table" "{{.EnvironmentName}}" {
  vpc_id = "${aws_vpc.{{.EnvironmentName}}.id}"

  tags = {
    Name                                        = "{{.EnvironmentName}}"
  }
}

resource "aws_route_table_association" "private-{{.Region}}a-{{.EnvironmentName}}" {
  subnet_id      = "${aws_subnet.{{.Region}}a-{{.EnvironmentName}}.id}"
  route_table_id = "${aws_route_table.private-{{.Region}}a-{{.EnvironmentName}}.id}"
}

resource "aws_route_table_association" "private-{{.Region}}b-{{.EnvironmentName}}" {
  subnet_id      = "${aws_subnet.{{.Region}}b-{{.EnvironmentName}}.id}"
  route_table_id = "${aws_route_table.private-{{.Region}}b-{{.EnvironmentName}}.id}"
}

resource "aws_route_table_association" "utility-{{.Region}}a-{{.EnvironmentName}}" {
  subnet_id      = "${aws_subnet.utility-{{.Region}}a-{{.EnvironmentName}}.id}"
  route_table_id = "${aws_route_table.{{.EnvironmentName}}.id}"
}

resource "aws_route_table_association" "utility-{{.Region}}b-{{.EnvironmentName}}" {
  subnet_id      = "${aws_subnet.utility-{{.Region}}b-{{.EnvironmentName}}.id}"
  route_table_id = "${aws_route_table.{{.EnvironmentName}}.id}"
}

resource "aws_subnet" "{{.Region}}a-{{.EnvironmentName}}" {
  vpc_id            = "${aws_vpc.{{.EnvironmentName}}.id}"
  cidr_block        = "172.20.32.0/19"
  availability_zone = "{{.Region}}a"

  tags = {
    Name                                        = "{{.Region}}a.{{.EnvironmentName}}"
    SubnetType                                  = "Private"
  }

  lifecycle {
    ignore_changes = ["tags"]
  }
}

resource "aws_subnet" "{{.Region}}b-{{.EnvironmentName}}" {
  vpc_id            = "${aws_vpc.{{.EnvironmentName}}.id}"
  cidr_block        = "172.20.64.0/19"
  availability_zone = "{{.Region}}b"

  tags = {
    Name                                        = "{{.Region}}b.{{.EnvironmentName}}"
    SubnetType                                  = "Private"
  }

  lifecycle {
    ignore_changes = ["tags"]
  }
}

resource "aws_subnet" "utility-{{.Region}}a-{{.EnvironmentName}}" {
  vpc_id            = "${aws_vpc.{{.EnvironmentName}}.id}"
  cidr_block        = "172.20.0.0/22"
  availability_zone = "{{.Region}}a"

  tags = {
    Name                                        = "utility-{{.Region}}a.{{.EnvironmentName}}"
    SubnetType                                  = "Utility"
  }

  lifecycle {
    ignore_changes = ["tags"]
  }
}

resource "aws_subnet" "utility-{{.Region}}b-{{.EnvironmentName}}" {
  vpc_id            = "${aws_vpc.{{.EnvironmentName}}.id}"
  cidr_block        = "172.20.4.0/22"
  availability_zone = "{{.Region}}b"

  tags = {
    Name                                        = "utility-{{.Region}}b.{{.EnvironmentName}}"
    SubnetType                                  = "Utility"
  }

  lifecycle {
    ignore_changes = ["tags"]
  }
}

resource "aws_vpc" "{{.EnvironmentName}}" {
  cidr_block           = "172.20.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name                                        = "{{.EnvironmentName}}"
  }
}

resource "aws_vpc_dhcp_options" "{{.EnvironmentName}}" {
  domain_name         = "{{.Region}}.compute.internal"
  domain_name_servers = ["AmazonProvidedDNS"]

  tags = {
    Name                                        = "{{.EnvironmentName}}"
  }
}

resource "aws_vpc_dhcp_options_association" "{{.EnvironmentName}}" {
  vpc_id          = "${aws_vpc.{{.EnvironmentName}}.id}"
  dhcp_options_id = "${aws_vpc_dhcp_options.{{.EnvironmentName}}.id}"
}
`

func RenderNetwork(config *model.Config) {
	terraformTemplate := parseNetworkTemplate(config.Spec.EnvironmentName, config.Spec.Region, config.Spec.ConfigBucket)
	util.WriteFile("./network.tf", terraformTemplate)
}

func parseNetworkTemplate(environmentName, region, configBucket string) []byte {
	var buf bytes.Buffer
	tmpl, err := template.New("networkTemplate").Parse(networkTemplate)
	util.CheckError(err)
	err = tmpl.Execute(&buf, struct {
		EnvironmentName string
		Region          string
		ConfigBucket    string
	}{environmentName, region, configBucket})
	util.CheckError(err)
	return buf.Bytes()
}
