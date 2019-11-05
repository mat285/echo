package vault

const (
	basePathSys         = "/v1/sys"
	basePathMounts      = basePathSys + "/mounts"
	basePathAuthMounts  = basePathSys + "/auth"
	basePathAuditMounts = basePathSys + "/audit"
)

// RenewLease renews and returns a new lease
func (c *Client) RenewLease(lease Lease, incrementSeconds int) (Lease, error) {
	req := c.NewRequest().AsPut().WithPathf("%s/renew", basePathSys)
	payload := map[string]interface{}{
		"lease_id":  lease.LeaseID,
		"increment": incrementSeconds,
	}
	var leaseResponse Lease
	return leaseResponse, vaultRequest(req.WithPostBodyAsJSON(payload), &leaseResponse)
}

// IsInitialized checks if the vault is initialized
func (c *Client) IsInitialized() (bool, error) {
	var initResponse vaultInitResponse
	return initResponse.Initialized, vaultRequest(c.NewRequest().AsGet().WithPathf("%s/init", basePathSys), &initResponse)
}

// Initialize initializes the vault and returns unseal keys and initial root token
func (c *Client) Initialize(numShares, threshold int) (Initialization, error) {
	var initialization Initialization
	req := c.NewRequest().AsPut().WithPathf("%s/init", basePathSys).WithPostBodyAsJSON(map[string]interface{}{
		"secret_shares":    numShares,
		"secret_threshold": threshold,
	})
	return initialization, vaultRequest(req, &initialization)
}

// IsSealed checks if the vault is sealed
func (c *Client) IsSealed() (bool, error) {
	var sealStatus SealStatus
	return sealStatus.Sealed, vaultRequest(c.NewRequest().AsGet().WithPathf("%s/seal-status", basePathSys), &sealStatus)
}

// Unseal unseals the vault with a specify unseal key
func (c *Client) Unseal(key string, reset, migrate bool) (SealStatus, error) {
	var sealStatus SealStatus
	req := c.NewRequest().AsPut().WithPathf("%s/unseal", basePathSys).WithPostBodyAsJSON(map[string]interface{}{
		"key":     key,
		"reset":   reset,
		"migrate": migrate,
	})
	return sealStatus, vaultRequest(req, &sealStatus)
}

// Mount mounts a backend with a config (can be nil)
func (c *Client) Mount(backend string, config map[string]interface{}) error {
	req := c.NewRequest().AsPost().WithPathf("%s/%s", basePathMounts, backend).WithPostBodyAsJSON(map[string]interface{}{
		"type":   backend,
		"config": config,
	})
	return vaultRequest(req, nil)
}

// TuneMount tunes the config of a backend
func (c *Client) TuneMount(backend string, defaultLeaseTTL, maxLeaseTTL int) error {
	req := c.NewRequest().AsPost().WithPathf("%s/%s/tune", basePathMounts, backend).WithPostBodyAsJSON(map[string]interface{}{
		"default_lease_ttl": defaultLeaseTTL,
		"max_lease_ttl":     maxLeaseTTL,
	})
	return vaultRequest(req, nil)
}

// MountAuth mounts an auth backend with a config (can be nil)
func (c *Client) MountAuth(backend string, config map[string]interface{}) error {
	req := c.NewRequest().AsPost().WithPathf("%s/%s", basePathAuthMounts, backend).WithPostBodyAsJSON(map[string]interface{}{
		"type":   backend,
		"config": config,
	})
	return vaultRequest(req, nil)
}

// MountAudit mounts an audit backend with a config (can be nil)
func (c *Client) MountAudit(backend string, config map[string]interface{}) error {
	req := c.NewRequest().AsPut().WithPathf("%s/%s", basePathAuditMounts, backend).WithPostBodyAsJSON(map[string]interface{}{
		"type":    backend,
		"options": config,
	})
	return vaultRequest(req, nil)
}

// Leader gets vault leader status
func (c *Client) Leader() (LeaderStatus, error) {
	var leaderStatus LeaderStatus
	return leaderStatus, vaultRequest(c.NewRequest().AsGet().WithPathf("%s/leader", basePathSys), &leaderStatus)
}
