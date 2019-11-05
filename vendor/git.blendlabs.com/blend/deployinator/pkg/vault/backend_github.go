package vault

import (
	"strings"

	"github.com/blend/go-sdk/exception"
)

const (
	basePathGithubAuth = "v1/auth/github"
)

// GithubLogin logs into vault through a github token
func (c *Client) GithubLogin(token string) (Token, error) {
	req := c.NewRequest().AsPost().WithPathf("%s/login", basePathGithubAuth).WithPostBodyAsJSON(map[string]interface{}{
		"token": token,
	})
	res := vaultAuthResponse{}
	err := vaultRequest(req, &res)
	if err != nil {
		return Token{}, exception.New(err)
	}
	return Token{
		ClientToken: res.Auth.ClientToken,
		Duration:    res.Auth.LeaseDuration,
		Renewable:   res.Auth.Renewable,
	}, nil
}

// GithubClient returns a vault client authenticated with the host through the given role
func GithubClient(host, token string) (*Client, error) {
	t, err := NewClient().WithHost(host).WithTLS().GithubLogin(token)
	if err != nil {
		return nil, exception.New(err)
	}
	return NewClient().WithHost(host).WithToken(t.ClientToken).WithTLS(), nil
}

// ListGithubTeamMappings lists team mappings
func (c *Client) ListGithubTeamMappings() ([]string, error) {
	var res vaultListResponse
	req := c.NewRequest().WithMethod("LIST").WithPathf("%s/map/teams", basePathGithubAuth)
	return res.Data.Keys, vaultRequest(req, &res)
}

// GetGithubTeamMapping gets a team mapping
func (c *Client) GetGithubTeamMapping(team string) ([]string, error) {
	var res vaultKeyValueResponse
	req := c.NewRequest().AsGet().WithPathf("%s/map/teams/%s", basePathGithubAuth, team)
	err := vaultRequest(req, &res)
	return strings.Split(res.Data.Value, ","), err
}

// MapGithubTeam maps team to policies
func (c *Client) MapGithubTeam(team string, policies []string) error {
	req := c.NewRequest().AsPost().WithPathf("%s/map/teams/%s", basePathGithubAuth, team).WithPostBodyAsJSON(map[string]string{
		"value": strings.Join(policies, ","),
	})
	return vaultRequest(req, nil)
}
