package types

import "time"

// Namespace is a kubernetes namespace
type Namespace struct {
	// Name is the name of the namespace
	Name string `json:"name" yaml:"name"`
	// CreatedBy is the target that created the namespace
	CreatedBy string `json:"createdBy" yaml:"createdBy"`
	// OwnedBy is the primary owner of the namespace
	OwnedBy string `json:"ownedBy" yaml:"ownedBy"`
	// Created is the time the namespace was created
	Created time.Time `json:"created" yaml:"created"`
}
