package model

import (
	"github.com/ghodss/yaml"
	"github.com/infinityworks/fk-infra/util"
	"io/ioutil"
)

func FetchConfig() *Config {
	var config Config
	configBytes, err := ioutil.ReadFile("./fk-infra.yml")
	util.CheckError(err)
	util.CheckError(yaml.Unmarshal(configBytes, &config))
	return &config
}

type Config struct {
	Spec Spec `json:"spec"`
}

type Kubernetes struct {
	Name          string `json:"name"`
	LoggingElasticSearchName string `json:"logging-elasticsearch-name"`
}

type Database struct {
	Name string `json:"name"`
}

type Queue struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type PeeringConnection struct {
	Name        string `json:"name"`
	PeerVpcCidr string `json:"peer-vpc-cidr"`
}

type ElasticSearch struct {
	Name string `json:"name"`
}

type Spec struct {
	EnvironmentName    string              `json:"environment-name"`
	Region             string              `json:"region"`
	EncryptionKey      string              `json:"encryption-key"`
	ConfigBucket       string              `json:"config-bucket"`
	Kubernetes         []Kubernetes        `json:"kubernetes,omitempty"`
	Databases          []Database          `json:"databases,omitempty"`
	Queues             []Queue             `json:"queues,omitempty"`
	ElasticSearch      []ElasticSearch     `json:"elasticsearch,omitempty"`
	PeeringConnections []PeeringConnection `json:"peering-connections,omitempty"`
}
