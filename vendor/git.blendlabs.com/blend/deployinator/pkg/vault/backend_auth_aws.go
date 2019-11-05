package vault

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"git.blendlabs.com/blend/deployinator/pkg/aws"
	"github.com/blend/go-sdk/exception"
)

const (
	basePathEC2     = "/v1/auth/aws-ec2"
	basePathAWSAuth = "v1/auth/aws"
)

// ConfigureEC2Client configures the root IAM credentials
func (c *Client) ConfigureEC2Client(accessKey, secretKey string) error {
	req := c.NewRequest().AsPost().WithPathf("%s/config/client", basePathEC2).WithPostBodyAsJSON(map[string]interface{}{
		"access_key": accessKey,
		"secret_key": secretKey,
	})
	return vaultRequest(req, nil)
}

// RegisterInstanceRole registers an instance role to allow logging in with vault
func (c *Client) RegisterInstanceRole(role, roleARN, period string, policies []string) error {
	return c.CreateAWSInstanceRole(role, AwsRole{
		"bound_iam_role_arn":        roleARN,
		"period":                    period,
		"policies":                  strings.Join(policies, ","),
		"disallow_reauthentication": false,
	})
}

// CreateIAMRole creates the iam role to allow logging in
func (c *Client) CreateIAMRole(role, roleARN, period string, policies []string) error {
	req := c.NewRequest().AsPost().WithPathf("%s/role/%s", basePathAWSAuth, role).WithPostBodyAsJSON(AwsRole{
		"auth_type":               "iam",
		"bound_iam_principal_arn": roleARN,
		"policies":                strings.Join(policies, ","),
		"period":                  period,
	})
	return vaultRequest(req, nil)
}

// ListAWSInstanceRoles list aws roles
func (c *Client) ListAWSInstanceRoles() ([]string, error) {
	var res vaultListResponse
	req := c.NewRequest().WithMethod("LIST").WithPathf("%s/roles", basePathEC2)
	return res.Data.Keys, vaultRequest(req, &res)
}

// GetAWSInstanceRole gets an aws role
func (c *Client) GetAWSInstanceRole(role string) (AwsRole, error) {
	var res vaultAwsRoleResponse
	req := c.NewRequest().AsGet().WithPathf("%s/role/%s", basePathEC2, role)
	return res.Data, vaultRequest(req, &res)
}

// CreateAWSInstanceRole creates an aws role from a payload
func (c *Client) CreateAWSInstanceRole(role string, payload AwsRole) error {
	req := c.NewRequest().AsPost().WithPathf("%s/role/%s", basePathEC2, role).WithPostBodyAsJSON(payload)
	return vaultRequest(req, nil)
}

// AWSIAMLogin logs into vault through aws iam
func (c *Client) AWSIAMLogin(role string, client *aws.AWS) (Token, error) {
	htp, err := client.GetCallerIdentitySignedRequest()
	if err != nil {
		return Token{}, err
	}
	js, err := json.Marshal(htp.Header)
	if err != nil {
		return Token{}, err
	}
	enc := base64.StdEncoding.EncodeToString(js)
	req := c.NewRequest().AsPost().WithPathf("%s/login", basePathAWSAuth).WithPostBodyAsJSON(map[string]interface{}{
		"role":                    role,
		"iam_http_request_method": "POST",
		"iam_request_url":         base64.StdEncoding.EncodeToString([]byte(aws.STSURL)),            //base64-encoding of https://sts.amazonaws.com/
		"iam_request_body":        base64.StdEncoding.EncodeToString([]byte(aws.STSGetIdenityBody)), //base64 encoding of Action=GetCallerIdentity&Version=2011-06-15
		"iam_request_headers":     enc,
	})
	res := vaultAuthResponse{}
	err = vaultRequest(req, &res)
	if err != nil {
		return Token{}, err
	}

	return Token{
		ClientToken: res.Auth.ClientToken,
		Duration:    res.Auth.LeaseDuration,
		Renewable:   res.Auth.Renewable,
	}, nil
}

// IAMClient returns a vault client authenticated with the host through the given role
func IAMClient(host, role string, client *aws.AWS) (*Client, error) {
	token, err := NewClient().WithHost(host).WithTLS().AWSIAMLogin(role, client)
	if err != nil {
		return nil, exception.New(err)
	}
	return NewClient().WithHost(host).WithToken(token.ClientToken).WithTLS(), nil
}
