package vault

import (
	"net/http"
	"time"

	request "github.com/blend/go-sdk/request"
)

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
