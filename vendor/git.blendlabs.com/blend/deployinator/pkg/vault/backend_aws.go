package vault

import (
	"fmt"
)

const (
	basePath       = "/v1/aws"
	basePathRoles  = basePath + "/roles"
	basePathCreds  = basePath + "/creds"
	basePathConfig = basePath + "/config"
)

// ConfigureRoot configures the root IAM credentials
func (c *Client) ConfigureRoot(accessKey, secretKey, region string) error {
	req := c.NewRequest().AsPost().WithPathf("%s/root", basePathConfig).WithPostBodyAsJSON(map[string]interface{}{
		"access_key": accessKey,
		"secret_key": secretKey,
		"region":     region,
	})
	return vaultRequest(req, nil)
}

// ConfigureLease configures the default lease duration for aws credentials
func (c *Client) ConfigureLease(ttl, ttlMax int) error {
	req := c.NewRequest().AsPost().WithPathf("%s/lease", basePathConfig).WithPostBodyAsJSON(map[string]interface{}{
		"lease":     fmt.Sprintf("%ds", ttl),
		"lease_max": fmt.Sprintf("%ds", ttlMax),
	})
	return vaultRequest(req, nil)
}

// CreateOrUpdateAwsRole creates or updates the specific vault aws role with either the policy arn or policy document json
func (c *Client) CreateOrUpdateAwsRole(role string, policy IAMPolicy) error {
	req := c.NewRequest().AsPost().WithPathf("%s/%s", basePathRoles, role).WithPostBodyAsJSON(policy)
	return vaultRequest(req, nil)
}

// ReadAwsRole returns 404 if role does not exist
func (c *Client) ReadAwsRole(role string) error {
	return vaultRequest(c.NewRequest().AsGet().WithPathf("%s/%s", basePathRoles, role), nil)
}

// IsAwsRolePresent returns true if this specific role in vault aws is already present
func (c *Client) IsAwsRolePresent(role string) (bool, error) {
	err := c.ReadAwsRole(role)
	return err == nil, IgnoreNotFoundError(err)
}

// DeleteAwsRole deletes the specified aws role
func (c *Client) DeleteAwsRole(role string) error {
	return vaultRequest(c.NewRequest().AsDelete().WithPathf("%s/%s", basePathRoles, role), nil)
}

// GenerateCredentialsForKey generates aws credentials. no sts token.
func (c *Client) GenerateCredentialsForKey(key string) (Lease, AwsCredentials, error) {
	req := c.NewRequest().AsGet().WithPathf("%s/%s", basePathCreds, key)
	var awsResponse vaultAwsResponse
	return awsResponse.Lease, awsResponse.Data, vaultRequest(req, &awsResponse)
}
