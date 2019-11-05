package vault

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	request "github.com/blendlabs/go-request"
	"github.com/blendlabs/go-util/env"
)

var (
	_default     *Client
	_defaultLock sync.Mutex

	// HACK
	maxRetries    = defaultMaxRetries
	retryInterval = defaultRetryInterval
)

// Default returns the default client.
func Default() *Client {
	return _default
}

// SetDefault sets the default client.
func SetDefault(client *Client) {
	_defaultLock.Lock()
	defer _defaultLock.Unlock()
	_default = client
}

// Client is a client for vault.
type Client struct {
	Token         string
	Host          string
	Scheme        Scheme
	TLSSkipVerify bool
	CACertPath    string
	rootCAPool    *x509.CertPool
	transport     *http.Transport
}

// NewRequest returns a new request.
func (c *Client) NewRequest() *request.Request {
	scheme := c.Scheme
	if len(scheme) == 0 {
		scheme = SchemeTLS
	}
	return request.New().
		WithScheme(string(scheme)).
		WithHeader(httpHeaderVaultToken, c.Token).
		WithHost(c.Host).
		WithTLSRootCAPool(c.ensureRootCAPool()).
		WithTransport(c.transport)
}

// NewClient returns a new vault client
func NewClient() *Client {
	return &Client{Scheme: SchemeTLS, transport: &http.Transport{}}
}

// NewClientFromEnv returns a new vault client with settings from environment
func NewClientFromEnv() (*Client, error) {
	if env.Env().HasVar(EnvVarVaultToken) && env.Env().HasVar(EnvVarVaultHost) {
		return &Client{
			Token:         env.Env().String(EnvVarVaultToken),
			Host:          env.Env().String(EnvVarVaultHost),
			Scheme:        SchemeTLS,
			TLSSkipVerify: env.Env().Bool(EnvVarVaultSkipVerify),
			CACertPath:    env.Env().String(EnvVarVaultCACert, "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"),
			transport:     &http.Transport{},
		}, nil
	}
	return nil, fmt.Errorf("vault: %s and %s required", EnvVarVaultHost, EnvVarVaultToken)
}

// WithToken sets the vault token for the client
func (c *Client) WithToken(token string) *Client {
	c.Token = token
	return c
}

// WithHost sets the vault host for the client
func (c *Client) WithHost(host string) *Client {
	c.Host = host
	return c
}

// WithScheme sets the vault host scheme
func (c *Client) WithScheme(scheme Scheme) *Client {
	c.Scheme = scheme
	return c
}

// WithTLS is a shortcut for WithScheme("https")
func (c *Client) WithTLS() *Client {
	return c.WithScheme(SchemeTLS)
}

// WithCACertPath sets the vault ca cert path
func (c *Client) WithCACertPath(path string) *Client {
	c.CACertPath = path
	return c
}

func (c *Client) ensureRootCAPool() *x509.CertPool {
	if c.rootCAPool == nil {
		certPool, err := x509.SystemCertPool()
		if err != nil {
			return nil
		}
		cert, err := ioutil.ReadFile(c.CACertPath)
		if os.IsNotExist(err) {
		} else if err != nil {
		} else if !certPool.AppendCertsFromPEM(cert) {
		}
		c.rootCAPool = certPool
	}
	return c.rootCAPool
}

// HACK: define the function using var to make it mockable
var vaultRequest = func(req *request.Request, successResponse interface{}) error {

	var errorResponse vaultErrorResponse
	if successResponse == nil {
		successResponse = new(struct{})
	}

	tries := 0
	var responseMeta *request.ResponseMeta
	requestSucceeded := func() (bool, error) {
		tries++
		meta, err := req.JSONWithErrorHandler(successResponse, &errorResponse)
		if err != nil {
			if tries > maxRetries { // retries = tries - 1, maxRetries 0 = no retries
				return false, err
			}

			return false, nil
		}
		responseMeta = meta
		return true, nil
	}
	if _, err := requestSucceeded(); err != nil {
		return err
	}
	return nil
}

type vaultResult struct {
	Data struct {
		Ciphertext string `json:"ciphertext"`
		Plaintext  string `json:"plaintext"`
		HMAC       string `json:"hmac"`
	} `json:"data"`
}

// AwsCredentials include an aws_key, secret_key, and optional security_token
type AwsCredentials struct {
	AccessKey     string `json:"access_key"`
	SecretKey     string `json:"secret_key"`
	SecurityToken string `json:"security_token"`
}

// AwsRole is an aws role payload
type AwsRole map[string]interface{}

// IAMPolicy refers to an iam policy by either an arn or policy document in json
type IAMPolicy struct {
	ARN  string `json:"arn,omitempty"`
	JSON string `json:"policy,omitempty"`
}

// Policy is a vault policy
type Policy struct {
	Name  string `json:"name"`
	Rules string `json:"rules"`
}

// Lease contains information regarding the lease of this credentials
type Lease struct {
	LeaseID   string `json:"lease_id"`
	Renewable bool   `json:"renewable"`
	Duration  int    `json:"lease_duration"`
}

// Token contains information of a vault token
type Token struct {
	ClientToken string `json:"client_token"`
	Duration    int    `json:"lease_duration"`
	Renewable   bool   `json:"renewable"`
}

// TokenRole contains information of a vault token role
type TokenRole struct {
	RoleName        string   `json:"name"`
	AllowedPolicies []string `json:"allowed_policies"`
	Renewable       bool     `json:"renewable"`
	Period          int      `json:"period"`
}

// Initialization contains information of a vault initialization
type Initialization struct {
	Keys      []string `json:"keys"`
	RootToken string   `json:"root_token"`
}

// SealStatus is a vault seal status
type SealStatus struct {
	Sealed    bool `json:"sealed"`
	Threshold int  `json:"t"`
	NumShares int  `json:"n"`
	Progress  int  `json:"progress"`
}

// LeaderStatus is a vault leader status
type LeaderStatus struct {
	HAEnabled            bool   `json:"ha_enabled"`
	IsSelf               bool   `json:"is_self"`
	LeaderAddress        string `json:"leader_address"`
	LeaderClusterAddress string `json:"leader_cluster_address"`
}

// TokenLookup is a response from vault for looking up tokens
type TokenLookup struct {
	Data TokenLookupData `json:"data"`
}

// TokenLookupData is the data inside a token lookup response
type TokenLookupData struct {
	ExpireTime time.Time `json:"expire_time"`
	Period     int       `json:"period"`
}

// LeaseLookupData is the data from a lease lookup call
type LeaseLookupData struct {
	ID              string     `json:"id"`
	IssueTime       time.Time  `json:"issue_time"`
	ExpireTime      time.Time  `json:"expire_time"`
	LastRenewalTime *time.Time `json:"last_renewal_time"`
	Renewable       bool       `json:"renewable"`
	TTL             int        `json:"ttl"`
}

// ListResponse is the result from vault list
type ListResponse struct {
	Keys []string `json:"keys"`
}

// KeyValueResponse is the key value response type
type KeyValueResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type vaultAwsResponse struct {
	Lease
	Data AwsCredentials `json:"data"`
}

type vaultAwsRoleResponse struct {
	Data AwsRole `json:"data"`
}

type vaultAuthResponse struct {
	Auth vaultAuth `json:"auth"`
}

type vaultAuth struct {
	Renewable     bool     `json:"renewable"`
	LeaseDuration int      `json:"lease_duration"`
	ClientToken   string   `json:"client_token"`
	Accessor      string   `json:"accessor"`
	Policies      []string `json:"policies"`
}

type vaultTokenResponse struct {
	Auth Token `json:"auth"`
}

type vaultTokenRoleResponse struct {
	Data TokenRole `json:"data"`
}

type vaultInitResponse struct {
	Initialized bool `json:"initialized"`
}

type vaultListResponse struct {
	Data ListResponse `json:"data"`
}

type vaultListPoliciesResponse struct {
	Policies []string `json:"policies"`
}

type vaultLeaseLookupResponse struct {
	Data LeaseLookupData `json:"data"`
}

type vaultKeyValueResponse struct {
	Data KeyValueResponse `json:"data"`
}

type vaultErrorResponse struct {
	Errors []string `json:"errors"`
}

type vaultError struct {
	error
	code int
}

type jsonRequest interface {
	JSONWithErrorHandler(interface{}, interface{}) (*request.ResponseMeta, error)
	Meta() *request.Meta
	Transport() *http.Transport
}

// Scheme is a scheme for a vault request
type Scheme string

const (
	// SchemeTLS is the scheme for tls
	SchemeTLS Scheme = "https"
	// SchemeDefault is the default scheme
	SchemeDefault Scheme = "http"
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
