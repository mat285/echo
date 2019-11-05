package types

import (
	core "git.blendlabs.com/blend/deployinator/pkg/core"
	uuid "github.com/blend/go-sdk/uuid"
)

// Trigger is a deploy trigger for a project
type Trigger struct {
	// UUID is a unique identifier for this trigger
	UUID string `json:"uuid"`
	// Github is a webhook trigger
	Github *GithubWebhookTrigger `json:"github"`
}

// GithubWebhookTrigger is a trigger for github webhooks
type GithubWebhookTrigger struct {
	// Repository is the github repo to listen for
	Repository string `json:"repository"`
	// Ref is the github branch to listen for
	Ref string `json:"branch"`
	// Events are the github events to listen for
	Events []string `json:"events"`
	//HookID is the id of the github webhook returned on creation
	HookID int64 `json:"hookID"`
	// Secret is the shared secret for hmac verification, stored in vault, not etcd
	Secret []byte `json:"-"`
}

// NewGithubTrigger creates a new github webhook trigger
func NewGithubTrigger(repo, ref string, events []string) (Trigger, error) {
	secret, err := core.Crypto.GenerateBase64Secret(64)
	if err != nil {
		return Trigger{}, err
	}
	return Trigger{
		UUID: uuid.V4().String(),
		Github: &GithubWebhookTrigger{
			Repository: repo,
			Ref:        ref,
			Events:     events,
			Secret:     []byte(secret),
		},
	}, nil
}
