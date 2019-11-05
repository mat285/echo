package types

// NotificationsConfig represents notifications options.
type NotificationsConfig struct {
	Emails       []string `json:"emails,omitempty" yaml:"emails,omitempty"`
	SlackChannel string   `json:"slackChannel,omitempty" yaml:"slackChannel,omitempty"`
	DatadogHost  string   `json:"datadogHost,omitempty" yaml:"datadogHost,omitempty"`
}

// Config returns relevant config values from the base config.
func (nc NotificationsConfig) Config() *Config {
	return &Config{
		Emails:       nc.Emails,
		SlackChannel: nc.SlackChannel,
		DatadogHost:  nc.DatadogHost,
	}
}
