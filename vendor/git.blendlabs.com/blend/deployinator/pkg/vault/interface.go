package vault

// TODO: define all interfaces and expose ClientInterface instead of the raw client

// Interface is the interface for vault client
type Interface interface {
	AuthAWSInterface
	AWSInterface
	GithubInterface
	PolicyInterface
	StorageInterface
	SystemInterface
	TokenInterface
	TransitInterface
}

// StorageInterface is the interface for vault storage backend
type StorageInterface interface {
	GetValue(key string) (string, error)
	GetValueInterface(key string) (interface{}, error)
	GetJSONValue(key string, out interface{}) error
	SetValue(key, value string) error
	SetValueInterface(key string, value interface{}) error
	DeleteValue(key string) error
	ListKeys(path string) ([]string, error)
	ListKeysRecursive(path string) ([]string, error)
	// TODO: define the rest
}

// AuthAWSInterface is the interface for vault aws auth backend
type AuthAWSInterface interface {
	ListAWSInstanceRoles() ([]string, error)
	GetAWSInstanceRole(role string) (AwsRole, error)
	CreateAWSInstanceRole(role string, payload AwsRole) error
}

// AWSInterface is the interface for vault aws secret backend
type AWSInterface interface {
}

// GithubInterface is the interface for vault github auth backend
type GithubInterface interface {
	ListGithubTeamMappings() ([]string, error)
	GetGithubTeamMapping(team string) ([]string, error)
	MapGithubTeam(team string, policies []string) error
}

// PolicyInterface is the interface for vault policy system backend
type PolicyInterface interface {
	CreateOrUpdatePolicy(name, rules string) error
	GetPolicy(name string) (*Policy, error)
	ListPolicies() ([]string, error)
	DeletePolicy(name string) error
}

// SystemInterface is the interface for vault system backend
type SystemInterface interface {
}

// TokenInterface is the interface for vault token auth backend
type TokenInterface interface {
}

// TransitInterface is the interface for vault transit secret backend
type TransitInterface interface {
}
