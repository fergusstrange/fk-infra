package kubernetes

import (
	"context"
	"github.com/ericchiang/k8s"
	"github.com/ghodss/yaml"
	"github.com/infinityworks/fk-infra/util"
	"io/ioutil"
	"log"
	"net/http"
	"os/user"
	"path/filepath"
	"reflect"
)

func CreateOrUpdate(req k8s.Resource) {
	client := newClient()
	err := client.Create(context.TODO(), req)

	if apiErr, ok := err.(*k8s.APIError); ok {
		if apiErr.Code == http.StatusConflict {
			err = client.Update(context.TODO(), req)
		}
	}

	if err != nil {
		log.Printf("Error applying %s %+v", reflect.TypeOf(req).String(), err)
	} else {
		log.Printf("Applied %s %s", reflect.TypeOf(req).String(), *req.GetMetadata().Name)
	}
}

func newClient() *k8s.Client {
	currentUser, err := user.Current()
	util.CheckError(err)
	data, err := ioutil.ReadFile(filepath.Join(currentUser.HomeDir, ".kube", "config"))
	var config k8s.Config
	util.CheckError(yaml.Unmarshal(data, &config))
	client, err := k8s.NewClient(&config)
	util.CheckError(err)
	return client
}
