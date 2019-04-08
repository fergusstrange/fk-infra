package cmd

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/infinityworks/fk-infra/aws"
	"github.com/infinityworks/fk-infra/crypto"
	"github.com/infinityworks/fk-infra/model"
	"github.com/infinityworks/fk-infra/util"
	"github.com/spf13/cobra"
)

const (
	FlagRegion          = "region"
	FlagEnvironmentName = "environment-name"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initiate this directory ready for use",
	Long:  "Creates basic config with defaults and necessary tooling ready to start creating a simple cluster and supporting infrastructure. Edit fk-infra.yml to change defaults",
	Run: func(cmd *cobra.Command, args []string) {
		envName, err := cmd.Flags().GetString(FlagEnvironmentName)
		util.CheckError(err)
		region, err := cmd.Flags().GetString(FlagRegion)
		util.CheckError(err)

		bucketLocation := aws.CreateBucket(envName, region)
		keyAlias := aws.CreateKmsKey(envName, region)

		configModel := model.Config{
			Spec: model.Spec{
				EnvironmentName: envName,
				Region:          region,
				EncryptionKey:   keyAlias,
				ConfigBucket:    bucketLocation,
				Kubernetes: []model.Kubernetes{{
					Name:                     gossipClusterFriendlyKubernetesName(envName),
					LoggingElasticSearchName: "logging",
				}},
				ElasticSearch: []model.ElasticSearch{{Name: "logging"},
				},
			},
		}

		configModelBytes, err := yaml.Marshal(&configModel)
		util.CheckError(err)
		util.WriteFile("./fk-infra.yml", configModelBytes)

		crypto.CreateOrValidateExistingKey()
	},
}

func gossipClusterFriendlyKubernetesName(envName string) string {
	return fmt.Sprintf("%s.k8s.local", envName)
}

func init() {
	initCmd.Flags().String(FlagEnvironmentName, "", "The name of the environment to initiate")
	initCmd.Flags().String(FlagRegion, "", "The region to create the environment")
	util.CheckError(initCmd.MarkFlagRequired(FlagEnvironmentName))
	util.CheckError(initCmd.MarkFlagRequired(FlagRegion))
	RootCmd.AddCommand(initCmd)
}
