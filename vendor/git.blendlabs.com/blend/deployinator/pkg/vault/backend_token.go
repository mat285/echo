package vault

import (
	"net/http"
	"strings"
	"time"
)

const (
	tokenBasePath      = "/v1/auth/token"
	tokenBasePathRoles = tokenBasePath + "/roles"
)

// CreateToken creates and returns a new token, optionally against a role if specified
func (c *Client) CreateToken(role string, ttl int, policies []string) (Token, error) {
	path := tokenBasePath + "/create"
	if len(role) > 0 {
		path += "/" + role
	}
	req := c.NewRequest().AsPost().WithPath(path)
	payload := map[string]interface{}{
		"policies":  policies,
		"ttl":       ttl,
		"renewable": true,
	}

	var tokenResponse vaultTokenResponse
	return tokenResponse.Auth, vaultRequest(req.WithPostBodyAsJSON(payload), &tokenResponse)
}

// CreatePeriodicToken creates and returns a new periodic token (requires root/sudo token)
func (c *Client) CreatePeriodicToken(periodSeconds int, policies []string) (Token, error) {
	req := c.NewRequest().AsPost().WithPath(tokenBasePath + "/create")
	payload := map[string]interface{}{
		"policies":  policies,
		"period":    periodSeconds,
		"renewable": true,
	}

	var tokenResponse vaultTokenResponse
	return tokenResponse.Auth, vaultRequest(req.WithPostBodyAsJSON(payload), &tokenResponse)
}

// LookupToken looks up a token by id
func (c *Client) LookupToken(token string) (TokenLookupData, error) {
	req := c.NewRequest().
		AsPost().
		WithPathf("%s/lookup", tokenBasePath).
		WithPostBodyAsJSON(map[string]string{
			"token": token,
		})
	var tokenLookup TokenLookup
	return tokenLookup.Data, vaultRequest(req, &tokenLookup)
}

// IsTokenPresent returns true of the token exists
func (c *Client) IsTokenPresent(token string) (bool, error) {
	if len(token) == 0 {
		return false, nil
	}
	tokenData, err := c.LookupToken(token)
	// vault returns `403 bad token` for expired or invalid token
	if ErrorCode(err) == http.StatusForbidden {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return time.Now().UTC().Before(tokenData.ExpireTime.UTC()), nil
}

// RenewSelfToken renews and returns the authenticated token, the increment is ignored for periodic token
func (c *Client) RenewSelfToken(incrementSeconds int) (Token, error) {
	req := c.NewRequest().AsPost().WithPathf("%s/renew-self", tokenBasePath)
	payload := map[string]interface{}{
		"increment": incrementSeconds, // This doesn't actually increment the ttl but rather reset the ttl to the increment value.
	}
	var tokenResponse vaultTokenResponse
	return tokenResponse.Auth, vaultRequest(req.WithPostBodyAsJSON(payload), &tokenResponse)
}

// RevokeToken revokes a token
func (c *Client) RevokeToken(token string) error {
	req := c.NewRequest().AsPost().WithPathf("%s/revoke", tokenBasePath)
	payload := map[string]interface{}{
		"token": token,
	}
	return vaultRequest(req.WithPostBodyAsJSON(payload), nil)
}

// RevokeOrphanToken revokes a token orphaning its children
func (c *Client) RevokeOrphanToken(token string) error {
	req := c.NewRequest().AsPost().WithPathf("%s/revoke-orphan", tokenBasePath)
	payload := map[string]interface{}{
		"token": token,
	}
	return vaultRequest(req.WithPostBodyAsJSON(payload), nil)
}

// CreateOrUpdateTokenRole creates and returns a new token role or updates the existing one
func (c *Client) CreateOrUpdateTokenRole(name string, policies []string, periodSeconds int) error {
	req := c.NewRequest().AsPost().WithPathf("%s/%s", tokenBasePathRoles, name)
	payload := map[string]interface{}{
		"name":             name,
		"allowed_policies": strings.Join(policies, ","),
		"renewable":        true,
		"period":           periodSeconds,
	}
	return vaultRequest(req.WithPostBodyAsJSON(payload), nil)
}

// ListTokenRoles lists all token roles
func (c *Client) ListTokenRoles() (ListResponse, error) {
	var listResponse vaultListResponse
	req := c.NewRequest().WithMethod("LIST").WithPath(tokenBasePathRoles)
	return listResponse.Data, vaultRequest(req, &listResponse)
}

// GetTokenRole gets a token role
func (c *Client) GetTokenRole(name string) (TokenRole, error) {
	var tokenRoleResponse vaultTokenRoleResponse
	req := c.NewRequest().AsGet().WithPathf("%s/%s", tokenBasePathRoles, name)
	return tokenRoleResponse.Data, vaultRequest(req, &tokenRoleResponse)
}

// DeleteTokenRole deletes a token role
func (c *Client) DeleteTokenRole(name string) error {
	return vaultRequest(c.NewRequest().AsDelete().WithPathf("%s/%s", tokenBasePathRoles, name), nil)
}
