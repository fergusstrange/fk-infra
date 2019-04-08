package templates

import (
	"encoding/json"
	"github.com/infinityworks/fk-infra/util"
)

type IamPolicy struct {
	Effect   string   `json:"Effect"`
	Action   []string `json:"Action"`
	Resource []string `json:"Resource"`
}

func NewAllowIamPolicy() *IamPolicy {
	return &IamPolicy{
		Effect: "Allow",
	}
}

func (policy *IamPolicy) Actions(actions ...string) *IamPolicy {
	policy.Action = actions
	return policy
}

func (policy *IamPolicy) Resources(resources ...string) *IamPolicy {
	policy.Resource = resources
	return policy
}

func IamPolicyJsonString(policies []*IamPolicy) string {
	if policies != nil {
		bytes, err := json.Marshal(policies)
		util.CheckError(err)
		return string(bytes)
	}
	return ""
}
