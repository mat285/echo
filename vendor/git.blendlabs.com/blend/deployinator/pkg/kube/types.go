package kube

// EtcdTLSConfig is a config for etcd tls
type EtcdTLSConfig struct {
	CA   string
	Key  string
	Cert string
}

// DatadogADConfigKey is a type for valid autodiscovery config key
type DatadogADConfigKey string

// Valid autodiscovery config keys
const (
	DatadogADCheckNames  DatadogADConfigKey = "check_names"
	DatadogADInstances   DatadogADConfigKey = "instances"
	DatadogADInitConfigs DatadogADConfigKey = "init_configs"
)
