package vault

import (
	"net/http"
	"time"
)

const (
	basePathLease = basePathSys + "/leases"
)

// LookupLease lookups the lease in vault and returns the information about it
func (c *Client) LookupLease(lease string) (LeaseLookupData, error) {
	req := c.NewRequest().AsPut().WithPathf("%s/%s", basePathLease, "lookup").WithPostBodyAsJSON(map[string]interface{}{
		"lease_id": lease,
	})
	var res vaultLeaseLookupResponse
	err := vaultRequest(req, &res)
	return res.Data, err
}

// IsLeasePresent returns if the lease is present and not expired in vault
func (c *Client) IsLeasePresent(id string) (bool, error) {
	if len(id) == 0 {
		return false, nil
	}

	lease, err := c.LookupLease(id)

	if ErrorCode(err) == http.StatusNotFound || ErrorCode(err) == http.StatusForbidden || ErrorCode(err) == http.StatusBadRequest {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return time.Now().UTC().Before(lease.ExpireTime.UTC()) && lease.TTL > 0, nil
}
