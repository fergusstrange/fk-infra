package terraform

import (
	"encoding/json"
	"fmt"
	"github.com/infinityworks/fk-infra/executable"
	"github.com/infinityworks/fk-infra/util"
	"github.com/mholt/archiver"
	"log"
	"os"
	"runtime"
	"strings"
)

type Output struct {
	Value string `json:"value"`
}

type Outputs struct {
	VpcId                 Output `json:"vpc_id"`
	VpcCidr               Output `json:"vpc_cidr_block"`
	SubnetA               Output `json:"subnet_a"`
	SubnetB               Output `json:"subnet_b"`
	UtilitySubnetA        Output `json:"subnet_utility-a"`
	UtilitySubnetB        Output `json:"subnet_utility-b"`
	MasterSecurityGroupId Output `json:"master_security_group_id"`
	WorkerSecurityGroupId Output `json:"worker_security_group_id"`

	outputBytes []byte
}

type DatabaseOutput struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	Password string `json:"password"`
}

type ElasticSearchOutput struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	Arn      string `json:"arn"`
}

func (outputs Outputs) PrivateSubnets() []string {
	return []string{outputs.SubnetA.Value, outputs.SubnetB.Value}
}

func (outputs Outputs) UtilitySubnets() []string {
	return []string{outputs.UtilitySubnetA.Value, outputs.UtilitySubnetB.Value}
}

func (outputs Outputs) DatabaseConfig() []DatabaseOutput {
	outputMap := make(map[string]interface{})
	util.CheckError(json.Unmarshal(outputs.outputBytes, &outputMap))
	var databaseOutputs []DatabaseOutput
	for key, val := range outputMap {
		if strings.HasPrefix(key, "database_output_") {
			var output DatabaseOutput
			embeddedJson := []byte(val.(map[string]interface{})["value"].(string))
			util.CheckError(json.Unmarshal(embeddedJson, &output))
			databaseOutputs = append(databaseOutputs, output)
		}
	}
	return databaseOutputs
}

func (outputs Outputs) ElasticSearchConfig() []ElasticSearchOutput {
	outputMap := make(map[string]interface{})
	util.CheckError(json.Unmarshal(outputs.outputBytes, &outputMap))
	var elasticSearchOutputs []ElasticSearchOutput
	for key, val := range outputMap {
		if strings.HasPrefix(key, "elasticsearch_output_") {
			var output ElasticSearchOutput
			embeddedJson := []byte(val.(map[string]interface{})["value"].(string))
			util.CheckError(json.Unmarshal(embeddedJson, &output))
			elasticSearchOutputs = append(elasticSearchOutputs, output)
		}
	}
	return elasticSearchOutputs
}

func PlanAndApply(approved bool) {
	ExecuteTerraform("init")
	ExecuteTerraform("plan")
	if approved {
		ExecuteTerraform("apply", "-auto-approve")
	}
}

func ExecuteTerraform(args ...string) []byte {
	return executable.CacheOrDownload(".fk-infra/terraform",
		func() string {
			downloadUrl := fmt.Sprintf("https://releases.hashicorp.com/terraform/0.11.11/terraform_0.11.11_%s_%s.zip", runtime.GOOS, runtime.GOARCH)
			log.Printf("Downloading terraform from %s", downloadUrl)
			return downloadUrl
		},
		func(tempBinaryLocation string) {
			util.CheckError(archiver.NewZip().Unarchive(tempBinaryLocation, ".fk-infra"))
			util.CheckError(os.Remove(tempBinaryLocation))
		},
		args...)
}

func fetchOutputsBytes() (output []byte) {
	defer func() {
		if recover() != nil {
			log.Println("Unable to fetch terraform outputs")
			output = []byte("{}")
		}
	}()

	return ExecuteTerraform("output", "-json")
}

func FetchTerraformOutputs() Outputs {
	var terraformOutputs Outputs
	outputBytes := fetchOutputsBytes()
	util.CheckError(json.Unmarshal(outputBytes, &terraformOutputs))
	terraformOutputs.outputBytes = outputBytes
	return terraformOutputs
}
