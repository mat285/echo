package types

import (
	"fmt"
	"time"

	core "git.blendlabs.com/blend/deployinator/pkg/core"
	uuid "github.com/blend/go-sdk/uuid"
)

// APIToken is an api token
type APIToken struct {
	ClientID     string     `json:"clientID" yaml:"clientID"`
	ClientSecret []byte     `json:"-" yaml:"-"`
	Label        string     `json:"label,omitempty" yaml:"label,omitempty"`
	Team         string     `json:"team,omitempty" yaml:"team,omitempty"`
	CreatedBy    string     `json:"createdBy" yaml:"createdBy"`
	CreatedAt    time.Time  `json:"createdAt" yaml:"createdAt"`
	ExpiredAt    *time.Time `json:"expiredAt,omitempty" yaml:"expiredAt,omitempty"`
}

// NewAPIToken creates a new API token
func NewAPIToken(user string) (*APIToken, error) {
	secret, err := core.Crypto.GenerateBase64Secret(48) // using 48 to get a 64 byte string
	if err != nil {
		return nil, err
	}
	if len(user) == 0 {
		return nil, fmt.Errorf("Must specify user for this token")
	}
	return &APIToken{
		ClientID:     uuid.V4().String(),
		ClientSecret: []byte(secret),
		CreatedBy:    user,
		CreatedAt:    time.Now().UTC(),
		ExpiredAt:    nil,
	}, nil
}

// NewAPITokenForTeam creates a token for the given team with the given name
func NewAPITokenForTeam(team, label, createdBy string) (*APIToken, error) {
	token, err := NewAPIToken(createdBy)
	if err != nil {
		return nil, err
	}
	token.Team = team
	token.Label = label
	return token, nil
}

// Target returns the target of the api token
func (t *APIToken) Target() string {
	target := UserTarget(t.CreatedBy)
	if t.IsTeamToken() {
		target = TeamTarget(t.Team)
	}
	return target
}

// IsTeamToken returns whether this token is a team token or not
func (t *APIToken) IsTeamToken() bool {
	return len(t.Team) > 0
}
