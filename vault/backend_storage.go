package vault

import (
	"encoding/json"
	"net/http"
	"strings"

	exception "github.com/blendlabs/go-exception"
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

// ErrorCode returns error code from vault, or 0 if not a vault error
func ErrorCode(err error) int {
	if vaultErr, ok := err.(*vaultError); ok {
		return vaultErr.code
	}
	return 0
}

// IsExistingMountError checks if an error is an existing mount error
func IsExistingMountError(err error) bool {
	err = ExceptionUnwrap(err)
	return ErrorCode(err) == http.StatusBadRequest &&
		(strings.Contains(err.Error(), "existing mount") || strings.Contains(err.Error(), "already in use"))
}

// IgnoreNotFoundError returns nil if the error is a not found error or the passed in error value otherwise
func IgnoreNotFoundError(err error) error {
	inner := ExceptionUnwrap(err)
	code := ErrorCode(inner)
	// check for regular 404 or 400 with a specific message
	// (e.g. a 400 error returned when calling the config endpoint of a transit key that doesn't exist)
	if code == http.StatusNotFound ||
		code == http.StatusBadRequest && (strings.Contains(inner.Error(), "no existing key") || strings.Contains(inner.Error(), "not found")) {
		return nil
	}
	return err
}

// NewNotFoundError returns vault not found error (for testing)
func NewNotFoundError(err error) error {
	return &vaultError{
		error: err,
		code:  http.StatusNotFound,
	}
}

// ExceptionUnwrap unwraps an exception.Ex object to gets the underlying error
func ExceptionUnwrap(err error) error {
	if ex, ok := err.(*exception.Exception); ok {
		err = ex.Inner()
	}
	return err
}
