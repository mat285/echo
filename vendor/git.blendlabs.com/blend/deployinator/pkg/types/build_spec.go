package types

// BuildSpec is the build spec that contains service, project or database
type BuildSpec struct {
	Service  *Service
	Project  *Project
	Database *Database
	Config   *Config
}
