package cmd

import (
	"github.com/infinityworks/fk-infra/crypto"
	"github.com/infinityworks/fk-infra/model"
	"github.com/infinityworks/fk-infra/templates"
	"github.com/infinityworks/fk-infra/terraform"
	"github.com/infinityworks/fk-infra/util"
	"github.com/spf13/cobra"
)

const (
	FlagApprove = "approve"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Initiate this directory ready for use",
	Long:  "Creates basic config with defaults and necessary tooling ready to start creating a simple cluster and supporting infrastructure. Edit fk-infra.yml to change defaults",
	Run: func(cmd *cobra.Command, args []string) {
		config := model.FetchConfig()
		approved, err := cmd.Flags().GetBool(FlagApprove)
		util.CheckError(err)

		crypto.DecryptKeys()

		templates.RenderNetwork(config)
		templates.RenderElasticSearch(config)
		templates.RenderDatabases(config)

		terraform.PlanAndApply(approved)

		terraformOutputs := terraform.FetchTerraformOutputs()

		templates.ApplyKubernetesClusters(config, terraformOutputs, approved)
	},
}

func init() {
	applyCmd.Flags().Bool(FlagApprove, false, "Approve the described infrastructure and apply it to the environment")
	RootCmd.AddCommand(applyCmd)
}
