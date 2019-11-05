package vault

import (
	"encoding/json"
	"path/filepath"
	"strings"

	exception "github.com/blend/go-sdk/exception"
)

// StorageResponse is the response to a vault storage request
type StorageResponse struct {
	RequestID     string      `json:"request_id"`
	LeaseID       string      `json:"lease_id"`
	Renewable     bool        `json:"renewable"`
	LeaseDuration int64       `json:"lease_duration"`
	Data          StorageData `json:"data"`
}

// StorageKeysResponse is the response to a vault keys list request
type StorageKeysResponse struct {
	RequestID     string          `json:"request_id"`
	LeaseID       string          `json:"lease_id"`
	Renewable     bool            `json:"renewable"`
	LeaseDuration int64           `json:"lease_duration"`
	Data          StorageKeysData `json:"data"`
}

// StorageData is the data of a response
type StorageData struct {
	Value interface{} `json:"value"`
}

// StorageKeysData is the data of list keys
type StorageKeysData struct {
	Keys []string `json:"keys"`
}

// ListKeys lists full keys at a path.
func (c *Client) ListKeys(path string) ([]string, error) {
	var res StorageKeysResponse
	err := IgnoreNotFoundError(vaultRequest(c.NewRequest().AsGet().WithPathf("/v1/%s", path).
		WithQueryString("list", "true"), &res))
	if err != nil {
		return nil, err
	}
	finalKeys := make([]string, len(res.Data.Keys))
	for index, key := range res.Data.Keys {
		finalKeys[index] = filepath.Join(path, key)
	}
	return finalKeys, nil
}

// ListKeysRecursive fully traverses a key path.
func (c *Client) ListKeysRecursive(path string) ([]string, error) {
	var res StorageKeysResponse
	err := IgnoreNotFoundError(vaultRequest(c.NewRequest().AsGet().WithPathf("/v1/%s", path).
		WithQueryString("list", "true"), &res))

	if err != nil {
		return nil, err
	}
	var totalKeys []string
	for _, key := range res.Data.Keys {
		if strings.HasSuffix(key, "/") {
			subKeys, err := c.ListKeysRecursive(filepath.Join(path, key))
			if err != nil {
				return nil, err
			}
			totalKeys = append(totalKeys, subKeys...)
		} else {
			totalKeys = append(totalKeys, filepath.Join(path, key))
		}
	}

	return totalKeys, nil
}

// GetValue gets a value from a storage key.
func (c *Client) GetValue(key string) (string, error) {
	res, err := c.GetValueInterface(key)
	if err != nil {
		return "", err
	}
	d, ok := res.(string)
	if !ok {
		if res == nil {
			return "", nil
		}
		return "", exception.New("Non string data. Please use the GetValueInterface function")
	}
	return d, nil
}

// GetValueInterface returns a value from storage as an interface
func (c *Client) GetValueInterface(key string) (interface{}, error) {
	var res StorageResponse
	err := vaultRequest(c.NewRequest().AsGet().WithPathf("/v1/%s", key), &res)
	if err != nil {
		return "", err
	}
	return res.Data.Value, nil
}

// GetJSONValue gets a json value from a storage key and parse it into `out`
func (c *Client) GetJSONValue(key string, out interface{}) error {
	js, err := c.GetValue(key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(js), out)
}

// SetValue sets a value in a storage key.
func (c *Client) SetValue(key, value string) error {
	return c.SetValueInterface(key, value)
}

// SetValueInterface sets a value a storage key
func (c *Client) SetValueInterface(key string, value interface{}) error {
	req := c.NewRequest().AsPost().WithPathf("/v1/%s", key).
		WithHeader("Content-Type", "application/json").
		WithPostBodyAsJSON(
			StorageData{
				Value: value,
			},
		)
	return vaultRequest(req, nil)
}

// DeleteValue deletes a value with a given key
func (c *Client) DeleteValue(key string) error {
	return vaultRequest(c.NewRequest().AsDelete().WithPathf("/v1/%s", key).
		WithHeader("Content-Type", "application/json"), nil)
}
