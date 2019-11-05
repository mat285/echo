package vault

import (
	"net/http"
	"strings"

	"git.blendlabs.com/blend/deployinator/pkg/core"
)

// ErrorCode returns error code from vault, or 0 if not a vault error
func ErrorCode(err error) int {
	if vaultErr, ok := err.(*vaultError); ok {
		return vaultErr.code
	}
	return 0
}

// IsExistingMountError checks if an error is an existing mount error
func IsExistingMountError(err error) bool {
	err = core.ExceptionUnwrap(err)
	return ErrorCode(err) == http.StatusBadRequest &&
		(strings.Contains(err.Error(), "existing mount") || strings.Contains(err.Error(), "already in use"))
}

// IgnoreNotFoundError returns nil if the error is a not found error or the passed in error value otherwise
func IgnoreNotFoundError(err error) error {
	inner := core.ExceptionUnwrap(err)
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
