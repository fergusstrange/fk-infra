package kubernetes

import (
	"github.com/ericchiang/k8s/apis/core/v1"
	v12 "github.com/ericchiang/k8s/apis/meta/v1"
	"github.com/infinityworks/fk-infra/terraform"
	"github.com/infinityworks/fk-infra/util"
)

func ApplyConfigMaps(outputs terraform.Outputs) {
	for _, elasticSearchConfig := range outputs.ElasticSearchConfig() {
		CreateOrUpdate(&v1.ConfigMap{
			Metadata: &v12.ObjectMeta{
				Name:      util.String(elasticSearchConfig.Name),
				Namespace: util.String("default"),
			},
			Data: map[string]string{
				"endpoint": elasticSearchConfig.Endpoint,
			},
		})
	}
}
