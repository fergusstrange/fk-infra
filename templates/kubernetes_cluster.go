package templates

import (
	"bytes"
	"fmt"
	"github.com/infinityworks/fk-infra/crypto"
	"github.com/infinityworks/fk-infra/kops"
	"github.com/infinityworks/fk-infra/kubernetes"
	"github.com/infinityworks/fk-infra/model"
	"github.com/infinityworks/fk-infra/terraform"
	"github.com/infinityworks/fk-infra/util"
	"text/template"
	"time"
)

const clusterTemplate = `
apiVersion: kops/v1alpha2
kind: Cluster
metadata:
  name: {{.ClusterName}}
spec:
  api:
    loadBalancer:
      type: Public
  authorization:
    rbac: {}
  channel: stable
  cloudProvider: aws
  configBase: s3://{{.ConfigBucket}}/kops
  etcdClusters:
  - etcdMembers:
    - instanceGroup: master-{{.Region}}a-1
      name: a-1
    - instanceGroup: master-{{.Region}}b-1
      name: b-1
    - instanceGroup: master-{{.Region}}a-2
      name: a-2
    name: main
  - etcdMembers:
    - instanceGroup: master-{{.Region}}a-1
      name: a-1
    - instanceGroup: master-{{.Region}}b-1
      name: b-1
    - instanceGroup: master-{{.Region}}a-2
      name: a-2
    name: events
  iam:
    allowContainerRegistry: true
    legacy: false
  additionalPolicies:
    {{if $.NodePolicies}}
    node: |
      {{$.NodePolicies}}
    {{end}}
    {{if $.MasterPolicies}}
    master: |
      {{$.MasterPolicies}}
    {{end}}
  kubelet:
    anonymousAuth: false
  kubernetesApiAccess:
  - 0.0.0.0/0
  kubernetesVersion: 1.11.9
  masterPublicName: api.{{.ClusterName}}
  networkCIDR: {{.VpcCidr}}
  networkID: {{.VpcId}}
  networking:
    weave:
      mtu: 8912
  nonMasqueradeCIDR: 100.64.0.0/10
  sshAccess:
  - 0.0.0.0/0
  subnets:
  - cidr: 172.20.32.0/19
    id: {{index .Subnets 0}}
    name: {{.Region}}a
    type: Private
    zone: {{.Region}}a
  - cidr: 172.20.64.0/19
    id: {{index .Subnets 1}}
    name: {{.Region}}b
    type: Private
    zone: {{.Region}}b
  - cidr: 172.20.0.0/22
    id: {{index .UtilitySubnets 0}}
    name: utility-{{.Region}}a
    type: Utility
    zone: {{.Region}}a
  - cidr: 172.20.4.0/22
    id: {{index .UtilitySubnets 1}}
    name: utility-{{.Region}}b
    type: Utility
    zone: {{.Region}}b
  topology:
    dns:
      type: Public
    masters: private
    nodes: private

---

apiVersion: kops/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: {{.ClusterName}}
  name: master-{{.Region}}a-1
spec:
  image: kope.io/k8s-1.11-debian-stretch-amd64-hvm-ebs-2018-08-17
  machineType: m4.large
  maxSize: 1
  minSize: 1
  nodeLabels:
    kops.k8s.io/instancegroup: master-{{.Region}}a-1
  role: Master
  additionalSecurityGroups:
  - {{.MasterSecurityGroupId}}
  subnets:
  - {{.Region}}a

---

apiVersion: kops/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: {{.ClusterName}}
  name: master-{{.Region}}a-2
spec:
  image: kope.io/k8s-1.11-debian-stretch-amd64-hvm-ebs-2018-08-17
  machineType: m4.large
  maxSize: 1
  minSize: 1
  nodeLabels:
    kops.k8s.io/instancegroup: master-{{.Region}}a-2
  role: Master
  additionalSecurityGroups:
  - {{.MasterSecurityGroupId}}
  subnets:
  - {{.Region}}a

---

apiVersion: kops/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: {{.ClusterName}}
  name: master-{{.Region}}b-1
spec:
  image: kope.io/k8s-1.11-debian-stretch-amd64-hvm-ebs-2018-08-17
  machineType: m4.large
  maxSize: 1
  minSize: 1
  nodeLabels:
    kops.k8s.io/instancegroup: master-{{.Region}}b-1
  role: Master
  additionalSecurityGroups:
  - {{.MasterSecurityGroupId}}
  subnets:
  - {{.Region}}b

---

apiVersion: kops/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: {{.ClusterName}}
  name: nodes
spec:
  image: kope.io/k8s-1.11-debian-stretch-amd64-hvm-ebs-2018-08-17
  machineType: m4.large
  maxSize: 1
  minSize: 1
  nodeLabels:
    kops.k8s.io/instancegroup: nodes
  role: Node
  additionalSecurityGroups:
  - {{.WorkerSecurityGroupId}}
  subnets:
  - {{.Region}}a
  - {{.Region}}b
`

type ClusterTemplate struct {
	ClusterName, Region, ConfigBucket, VpcId,
	VpcCidr, MasterSecurityGroupId, WorkerSecurityGroupId, MasterPolicies,
	NodePolicies string
	Subnets, UtilitySubnets []string
}

func ApplyKubernetesClusters(config *model.Config, outputs terraform.Outputs, approved bool) {
	if baseVPCExists(outputs) {
		configBucket := config.Spec.ConfigBucket

		elasticSearchMasterPolicy, elasticSearchNodePolicy := masterAndNodeIamPolicies(outputs)

		for _, kubernetesCluster := range config.Spec.Kubernetes {
			clusterName := kubernetesCluster.Name

			clusterTemplate := parseClusterTemplate(
				clusterName,
				elasticSearchMasterPolicy,
				elasticSearchNodePolicy,
				config,
				outputs)

			kopsFileName := kopsTemplateFilename(clusterName)
			util.WriteFile(kopsFileName, clusterTemplate)

			kops.ExecuteKops(kopsStateFlag(configBucket), "replace", "-f", kopsFileName, "--force")
			kops.ExecuteKops(kopsStateFlag(configBucket), "create", "secret", kopsClusterNameFlag(clusterName), "sshpublickey", "admin", "-i", crypto.PublicKeyFile)
			kops.ExecuteKops(kopsUpdateCluster(configBucket, clusterName, approved)...)

			if approved {
				validateCluster(configBucket, clusterName)

				kubernetes.ApplyServices(outputs)
				kubernetes.ApplyConfigMaps(outputs)
				kubernetes.ApplySecrets(outputs)
				applyLogging(kubernetesCluster, outputs, config)
			}
		}
	}
}

func masterAndNodeIamPolicies(outputs terraform.Outputs) (masterPolicies string, nodePolicies string) {
	elasticSearchMasterPolicies, elasticSearchNodePolicies := elasticSearchIamPolicies(outputs)
	allNodePolicies := flattenIamPolicies(elasticSearchNodePolicies, route53NodePolicies())
	return IamPolicyJsonString(elasticSearchMasterPolicies), IamPolicyJsonString(allNodePolicies)
}

func flattenIamPolicies(policiesToFlatten ...[]*IamPolicy) []*IamPolicy {
	iamPolicies := make([]*IamPolicy, 0)
	for _, policies := range policiesToFlatten {
		for _, policy := range policies {
			iamPolicies = append(iamPolicies, policy)
		}
	}
	return iamPolicies
}

func route53NodePolicies() []*IamPolicy {
	return []*IamPolicy{NewAllowIamPolicy().
		Actions("route53:ListHostedZones",
			"route53:ListResourceRecordSets",
			"route53:ChangeResourceRecordSets",
			"route53:ListHostedZonesByName",
			"route53:GetChange").
		Resources("*")}
}

func elasticSearchIamPolicies(outputs terraform.Outputs) (masterPolicies []*IamPolicy, nodePolicies []*IamPolicy) {
	var masterIamPolicies []*IamPolicy
	var nodeIamPolicies []*IamPolicy
	for _, elasticSearchCluster := range outputs.ElasticSearchConfig() {
		iamPolicy := NewAllowIamPolicy().
			Actions("es:*").
			Resources(fmt.Sprintf("%s/*", elasticSearchCluster.Arn))
		masterIamPolicies = append(masterIamPolicies, iamPolicy)
		nodeIamPolicies = append(nodeIamPolicies, iamPolicy)
	}
	return masterIamPolicies, nodeIamPolicies
}

func applyLogging(kubernetesCluster model.Kubernetes, outputs terraform.Outputs, config *model.Config) {
	if kubernetesCluster.LoggingElasticSearchName != "" {
		for _, elasticSearchCluster := range outputs.ElasticSearchConfig() {
			if kubernetesCluster.LoggingElasticSearchName == elasticSearchCluster.Name {
				kubernetes.ApplyFluentBitLogging(elasticSearchCluster.Endpoint, config.Spec.Region)
				break
			}
		}
	}
}

func parseClusterTemplate(clusterName, masterPolicy, nodePolicy string, config *model.Config, outputs terraform.Outputs) []byte {
	var buf bytes.Buffer
	tmpl, err := template.New("clusterTemplate").Parse(clusterTemplate)
	util.CheckError(err)
	err = tmpl.Execute(&buf, ClusterTemplate{
		ClusterName:           clusterName,
		Region:                config.Spec.Region,
		ConfigBucket:          config.Spec.ConfigBucket,
		VpcId:                 outputs.VpcId.Value,
		VpcCidr:               outputs.VpcCidr.Value,
		MasterSecurityGroupId: outputs.MasterSecurityGroupId.Value,
		WorkerSecurityGroupId: outputs.WorkerSecurityGroupId.Value,
		Subnets:               outputs.PrivateSubnets(),
		UtilitySubnets:        outputs.UtilitySubnets(),
		MasterPolicies:        masterPolicy,
		NodePolicies:          nodePolicy,
	})
	util.CheckError(err)
	return buf.Bytes()
}

func validateCluster(configBucket string, clusterName string) {
	var clusterIsReady = func() (ready bool) {
		ready = true
		defer func() {
			if recover() != nil {
				ready = false
				time.Sleep(15 * time.Second)
			}
		}()
		kops.ExecuteKops(kopsStateFlag(configBucket), "validate", "cluster", kopsClusterNameFlag(clusterName))
		return true
	}

	inTenMins := time.Now().Add(10 * time.Minute)
	for time.Now().Before(inTenMins) {
		if clusterIsReady() {
			return
		}
	}
}

func baseVPCExists(terraformOutputs terraform.Outputs) bool {
	return terraformOutputs.VpcId.Value != ""
}

func kopsTemplateFilename(clusterName string) string {
	return fmt.Sprintf("./%s.yml", clusterName)
}

func kopsUpdateCluster(configBucket, clusterName string, approved bool) []string {
	args := []string{kopsStateFlag(configBucket), "update", "cluster", kopsClusterNameFlag(clusterName)}

	if approved {
		args = append(args, "--yes")
	}

	return args
}

func kopsClusterNameFlag(clusterName string) string {
	return fmt.Sprintf("--name=%s", clusterName)
}

func kopsStateFlag(configBucket string) string {
	return fmt.Sprintf("--state=s3://%s", kopsConfigBucket(configBucket))
}

func kopsConfigBucket(configBucket string) string {
	return fmt.Sprintf("%s/kops", configBucket)
}
