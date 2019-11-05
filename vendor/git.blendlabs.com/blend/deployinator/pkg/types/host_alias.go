package types

// HostAlias represents an entry in a hosts file.
type HostAlias struct {
	IP        string   `json:"ip" yaml:"ip"`
	HostNames []string `json:"hostnames" yaml:"hostnames"`
}
