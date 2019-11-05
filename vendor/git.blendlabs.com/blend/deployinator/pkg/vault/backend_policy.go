package vault

const (
	policyBasePath = "v1/sys/policy"
)

// CreateOrUpdatePolicy creates or updates a policy
func (c *Client) CreateOrUpdatePolicy(name string, rules string) error {
	req := c.NewRequest().AsPut().WithPathf("%s/%s", policyBasePath, name)
	payload := map[string]interface{}{
		"rules": rules,
	}
	return vaultRequest(req.WithPostBodyAsJSON(payload), nil)
}

// GetPolicy returns the policy from the name
func (c *Client) GetPolicy(name string) (*Policy, error) {
	req := c.NewRequest().AsGet().WithPathf("%s/%s", policyBasePath, name)
	var res Policy
	err := vaultRequest(req, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// ListPolicies returns all policies in vault
func (c *Client) ListPolicies() ([]string, error) {
	req := c.NewRequest().AsGet().WithPath(policyBasePath)
	var res vaultListPoliciesResponse
	return res.Policies, vaultRequest(req, &res)
}

// DeletePolicy deletes a policy
func (c *Client) DeletePolicy(name string) error {
	req := c.NewRequest().AsDelete().WithPathf("%s/%s", policyBasePath, name)
	return vaultRequest(req, nil)
}
