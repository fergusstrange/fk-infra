package kubernetes

import (
	"github.com/ericchiang/k8s/apis/core/v1"
	v12 "github.com/ericchiang/k8s/apis/meta/v1"
	"github.com/infinityworks/fk-infra/terraform"
	"github.com/infinityworks/fk-infra/util"
)

func ApplySecrets(outputs terraform.Outputs) {
	for _, databaseConfig := range outputs.DatabaseConfig() {
		CreateOrUpdate(&v1.Secret{
			Metadata: &v12.ObjectMeta{
				Name:      util.String(databaseConfig.Name),
				Namespace: util.String("default"),
			},
			Type: util.String("Opaque"),
			Data: map[string][]byte{
				"schema":   []byte(databaseConfig.Name),
				"endpoint": []byte(databaseConfig.Endpoint),
				"password": []byte(databaseConfig.Password),
			},
		})
	}
}
