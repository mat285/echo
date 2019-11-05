package vault

import (
	"time"
)

const (
	// EnvVarVaultToken is an environment variable with the vault token
	EnvVarVaultToken = "VAULT_TOKEN"
	// EnvVarVaultHost is an environment variable with the vault host
	EnvVarVaultHost = "VAULT_HOST"
	// EnvVarVaultAddr is an environment variable with the vault address (protocol + host)
	EnvVarVaultAddr = "VAULT_ADDR"
	// EnvVarVaultCACert is an environment variable with the vault ca cert path
	EnvVarVaultCACert = "VAULT_CACERT"
	// EnvVarVaultSkipVerify is an environment variable for skipping tls certificate verification
	EnvVarVaultSkipVerify = "VAULT_SKIP_VERIFY"

	// EnvVarLeaseID is the env var for the lease ID
	EnvVarLeaseID = "LEASE_ID"

	// EnvVarVaultAWSAccessKeyID is the aws access key for vault
	EnvVarVaultAWSAccessKeyID = "VAULT_AWS_ACCESS_KEY_ID"
	// EnvVarVaultAWSSecretAccessKey is the aws secret access key for vault
	EnvVarVaultAWSSecretAccessKey = "VAULT_AWS_SECRET_ACCESS_KEY"

	// PolicyDefault is the default vault policy
	PolicyDefault = "default"

	// InitialLeaseDurationSeconds initial duration for vault leases before they expire (ttl)
	InitialLeaseDurationSeconds = int(45 * time.Minute / time.Second)
	// LeaseDurationSeconds target duration for vault leases before they expire (ttl)
	LeaseDurationSeconds = int(24 * time.Hour / time.Second)
	// LeaseMaxSeconds maximum duration of vault leases before they expire (ttl), they cannot be renewed beyond that
	LeaseMaxSeconds = int(32 * 24 * time.Hour / time.Second)
	// TokenDurationSeconds target duration for vault tokens before they expire (ttl)
	TokenDurationSeconds = int(24 * time.Hour / time.Second)

	// BackendAWS is the AWS backend type and mount path
	BackendAWS = "aws"
	// BackendTransit is the transit backend type and mount path
	BackendTransit = "transit"
	// BackendAuthAWS is the AWS auth backend type and mount path
	BackendAuthAWS = "aws"
	// BackendAuthEC2 is the AWS EC2 auth backend type and mount path
	// TODO: use `aws` instead because `aws-ec2` is obsoleted. https://github.com/hashicorp/vault/pull/2441
	BackendAuthEC2 = "aws-ec2"
	// BackendAuditFile is an audit backend type and mount path
	BackendAuditFile = "file"

	// TransitPrefix is the prefix for transit-encrypted data
	TransitPrefix = "vault:v1:"

	httpHeaderVaultToken = "X-Vault-Token"

	defaultRetryInterval = 10 * time.Second
	defaultMaxRetries    = 5
)
