package types

import (
	"time"

	"k8s.io/apimachinery/pkg/util/rand"
)

// Database is the metadata for a database in deployinator
type Database struct {
	// Name is the name of the database
	Name string `json:"name" yaml:"name"`
	// DBStorage is the amount of storage the database is configured to use
	DBStorage int `json:"dbStorage,string" yaml:"dbStorage,string"`

	// Created is the timestamp when the service was created.
	Created time.Time `json:"created" yaml:"created"`
	// CreatedBy is the username that created the service.
	CreatedBy string `json:"createdBy" yaml:"createdBy"`

	// Current represents the currently provisioned config.
	Current Config `json:"current" yaml:"current"`

	// Last represents the last config for any build type.
	Last *Config `json:"last,omitempty" yaml:"last,omitempty"`

	// Env is the environment of the database
	Env string `json:"env" yaml:"env"`

	// DBRelaunch is the flag to indicate that database is being restored
	DBRelaunch bool `json:"dbRestore" yaml:"dbRestore"`
	// DBRestore is the flag to indicate that database is being restored
	DBRestore bool `json:"dbRelaunch" yaml:"dbRelaunch"`
	// DBBackupFileName is the name of the backup file to restore from if restoring
	DBBackupFileName string `json:"dbBackupFileName" yaml:"dbBackupFileName"`
}

// Config returns the config for this database
func (d *Database) Config() *Config {
	c := &Config{
		DatabaseName:     d.Name,
		DBStorage:        d.DBStorage,
		Resources:        d.Current.Resources,
		DeployID:         rand.String(RandomStringLength),
		Schedule:         d.Current.Schedule,
		Accessibility:    d.Current.Accessibility,
		DBRelaunch:       d.DBRelaunch,
		DBRestore:        d.DBRestore,
		DBBackupFileName: d.DBBackupFileName,
	}
	return c
}
