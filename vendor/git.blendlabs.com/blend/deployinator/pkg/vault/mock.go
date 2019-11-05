package vault

import (
	"encoding/json"
	"sync"
)

var (
	defaultVaultRequest = vaultRequest
	mockVaultMutex      = new(sync.Mutex)
)

// MockVault mocks vault
func MockVault() (chan interface{}, chan error) {
	resps, errs := make(chan interface{}, 20), make(chan error, 20)
	// unlock in unmock function
	mockVaultMutex.Lock()
	vaultRequest = func(req jsonRequest, successResponse interface{}) error {
		// pull responses
		resp := <-resps
		err := <-errs
		if err != nil {
			return err
		}
		if resp == nil {
			return nil
		}
		// marshal the response
		bytes, err := json.Marshal(resp)
		if err != nil {
			return err
		}
		// unmarshal the response into successResponse
		return json.Unmarshal(bytes, successResponse)
	}
	SetDefault(&Client{})
	return resps, errs
}

// UnmockVault unmocks vault
func UnmockVault() {
	defer mockVaultMutex.Unlock()
	vaultRequest = defaultVaultRequest
	SetDefault(nil)
}
