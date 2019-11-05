package vault

import (
	"encoding/base64"
	"encoding/json"
	"log"
)

// CreateTransitKey creates a new transit key.
func (c *Client) CreateTransitKey(key string) error {
	return vaultRequest(c.NewRequest().AsPost().WithPathf("/v1/transit/keys/%s", key).WithPostBodyAsJSON(map[string]interface{}{
		"type":    "aes256-gcm96",
		"derived": true,
	}), nil)
}

// ConfigureTransitKey configures a transit key.
func (c *Client) ConfigureTransitKey(key string, config map[string]interface{}) error {
	return vaultRequest(c.NewRequest().AsPost().WithPathf("/v1/transit/keys/%s/config", key).WithPostBodyAsJSON(config), nil)
}

// DeleteTransitKey deletes a transit key.
func (c *Client) DeleteTransitKey(key string) error {
	return vaultRequest(c.NewRequest().AsDelete().WithPathf("/v1/transit/keys/%s", key), nil)
}

// ReadTransitKey reads a transit key.
func (c *Client) ReadTransitKey(key string) error {
	return vaultRequest(c.NewRequest().AsGet().WithPathf("/v1/transit/keys/%s", key), nil)
}

// ListTransitKeys lists all transit keys.
func (c *Client) ListTransitKeys() ([]string, error) {
	var result vaultListResponse
	return result.Data.Keys, vaultRequest(c.NewRequest().WithMethod("LIST").WithPath("/v1/transit/keys"), &result)
}

// IsTransitKeyPresent returns true if the key is present in vault
func (c *Client) IsTransitKeyPresent(key string) (bool, error) {
	err := c.ReadTransitKey(key)
	return err == nil, IgnoreNotFoundError(err)
}

// CreateTransitKeyIfNotExists create transit key if it does not exist
func (c *Client) CreateTransitKeyIfNotExists(key string) error {
	keyIsPresent, err := c.IsTransitKeyPresent(key)
	if err != nil {
		return err
	}
	if keyIsPresent {
		return nil
	}
	return c.CreateTransitKey(key)
}

// RotateTransitKey rotates a transit key.
func (c *Client) RotateTransitKey(key string) error {
	return vaultRequest(c.NewRequest().AsPost().WithPathf("/v1/transit/keys/%s/rotate", key), nil)
}

// TransitEncrypt encrypts a given set of data.
func (c *Client) TransitEncrypt(key string, context map[string]interface{}, data []byte) (string, error) {
	log.Printf("Encrypting at key %s with context %v", key, context)
	req := c.NewRequest().AsPost().WithPathf("/v1/transit/encrypt/%s", key)
	payload := map[string]interface{}{
		"plaintext": base64.StdEncoding.EncodeToString(data),
	}
	if context != nil {
		contextJSON, _ := json.Marshal(context)
		contextEncoded := base64.StdEncoding.EncodeToString(contextJSON)
		payload["context"] = contextEncoded
	}

	var encryptionResult vaultResult
	err := vaultRequest(req.WithPostBodyAsJSON(payload), &encryptionResult)
	if err != nil {
		return "", err
	}
	log.Printf("transit encrypt: Encrypted for %s\n", key)

	return encryptionResult.Data.Ciphertext, nil
}

// TransitDecrypt decrypts a given set of data.
func (c *Client) TransitDecrypt(key string, context map[string]interface{}, ciphertext string) ([]byte, error) {
	log.Printf("Decrypting at key %s with context %v", key, context)
	req := c.NewRequest().AsPost().WithPathf("/v1/transit/decrypt/%s", key)
	payload := map[string]interface{}{
		"ciphertext": ciphertext,
	}
	if context != nil {
		contextJSON, err := json.Marshal(context)
		if err != nil {
			return nil, err
		}
		contextEncoded := base64.StdEncoding.EncodeToString(contextJSON)
		payload["context"] = contextEncoded
	}

	var encryptionResult vaultResult
	err := vaultRequest(req.WithPostBodyAsJSON(payload), &encryptionResult)
	if err != nil {
		return nil, err
	}
	log.Printf("transit decrypt: Decrypted for %s\n", key)
	return base64.StdEncoding.DecodeString(encryptionResult.Data.Plaintext)
}

// TransitHMAC computes the hmac hash of a set of data with a given key.
func (c *Client) TransitHMAC(key string, data []byte) ([]byte, error) {
	req := c.NewRequest().AsPost().WithPathf("/v1/transit/hmac/%s", key)
	payload := map[string]interface{}{
		"input": base64.StdEncoding.EncodeToString(data),
	}

	var encryptionResult vaultResult
	err := vaultRequest(req.WithPostBodyAsJSON(payload), &encryptionResult)
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(encryptionResult.Data.HMAC)
}
