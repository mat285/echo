package types

import (
	"fmt"
	"time"

	uuid "github.com/blend/go-sdk/uuid"
)

const (
	// UserAutodeploy is a special user that kicks off automatic deploys.
	UserAutodeploy = "auto_deploy"
	// UserSystem is the internal super user.
	UserSystem = "system"
)

// User represents a deployinator user.
type User struct {
	Username   string    `json:"username" yaml:"username"`
	CreatedAt  time.Time `json:"created" yaml:"created"`
	FirstName  string    `json:"firstName" yaml:"firstName"`
	LastName   string    `json:"lastName" yaml:"lastName"`
	Email      string    `json:"email" yaml:"email"`
	PictureURL string    `json:"pictureURL" yaml:"pictureURL"`
	IsSystem   bool      `json:"isSystem" yaml:"isSystem"`
}

// NewTestUser creates a new test user.
func NewTestUser() User {
	return User{
		Username:  uuid.V4().String(),
		FirstName: uuid.V4().String(),
		LastName:  uuid.V4().String(),
		Email:     fmt.Sprintf("%s@%s", uuid.V4().String(), uuid.V4().String()),
		CreatedAt: time.Now().UTC(),
	}
}

// NewSystemUser creates a new test user.
func NewSystemUser() User {
	return User{
		Username:  UserSystem,
		FirstName: "System",
		LastName:  "User",
		CreatedAt: time.Now().UTC(),
		IsSystem:  true,
	}
}
