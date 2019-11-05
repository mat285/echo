package types

import (
	"fmt"

	"github.com/blend/go-sdk/exception"
)

// StorageConfig is a config for persistent storage
type StorageConfig struct {
	// Capacity is the capacity of the requested storage in gigabytes
	Capacity string `json:"capacity,omitempty" yaml:"capacity,omitempty"`
	// MountPath is the path to mount the persistent volume at
	MountPath string `json:"mountPath,omitempty" yaml:"mountPath,omitempty"`
}

// Validate validates the storage config
func (sc StorageConfig) Validate() error {
	cap := sc.GetCapacity()
	path := sc.GetMountPath()
	if len(path) == 0 {
		return exception.New(fmt.Sprintf("No mount path set for storage"))
	}
	if len(cap) == 0 {
		return exception.New(fmt.Sprintf("No capacity set for storage"))
	}
	return nil
}

// GetCapacity returns the capacity of this storage config
func (sc StorageConfig) GetCapacity(defaults ...string) string {
	return thisOrThatOrDefault(sc.Capacity, "", defaults...)
}

// GetMountPath gets the mount path of this storage config
func (sc StorageConfig) GetMountPath(defaults ...string) string {
	return thisOrThatOrDefault(sc.MountPath, "", defaults...)
}
