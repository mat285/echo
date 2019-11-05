package types

import (
	"fmt"
	"time"

	uuid "github.com/blend/go-sdk/uuid"
)

// Team represents a group of users.
type Team struct {
	Name      string    `json:"name" yaml:"name"`
	CreatedBy string    `json:"createdBy" yaml:"createdBy"`
	CreatedAt time.Time `json:"createdAt" yaml:"createdAt"`
}

// TeamMember represents a user on a team.
type TeamMember struct {
	TeamName string    `json:"teamName" yaml:"teamName"`
	Username string    `json:"username" yaml:"username"`
	AddedBy  string    `json:"addedBy" yaml:"addedBy"`
	AddedAt  time.Time `json:"addedAt" yaml:"addedAt"`
}

// NewTestTeam creates a test team created by system.
func NewTestTeam() Team {
	return Team{
		Name:      uuid.V4().String(),
		CreatedAt: time.Now().UTC(),
		CreatedBy: UserSystem,
	}
}

func (t Team) String() string {
	return fmt.Sprintf("Name: `%s`, CreatedBy: `%s`, CreatedAt: `%s`", t.Name, t.CreatedBy, t.CreatedAt.UTC())
}
