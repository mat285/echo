package vault

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"

	"git.blendlabs.com/blend/deployinator/pkg/kube"
	"git.blendlabs.com/blend/deployinator/pkg/logging"
	"github.com/blend/go-sdk/env"
	exception "github.com/blend/go-sdk/exception"
	request "github.com/blend/go-sdk/request"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	_default     *Client
	_defaultLock sync.Mutex

	// HACK
	maxRetries    = defaultMaxRetries
	retryInterval = defaultRetryInterval
)

// Default returns the default client.
func Default() *Client {
	return _default
}

// SetDefault sets the default client.
func SetDefault(client *Client) {
	_defaultLock.Lock()
	defer _defaultLock.Unlock()
	_default = client
}

// Client is a client for vault.
type Client struct {
	Token         string
	Host          string
	Scheme        Scheme
	TLSSkipVerify bool
	CACertPath    string
	rootCAPool    *x509.CertPool
	transport     *http.Transport
}

// NewRequest returns a new request.
func (c *Client) NewRequest() *request.Request {
	scheme := c.Scheme
	if len(scheme) == 0 {
		scheme = SchemeTLS
	}
	return request.New().
		WithScheme(string(scheme)).
		WithHeader(httpHeaderVaultToken, c.Token).
		WithHost(c.Host).
		WithTLSSkipVerify(c.TLSSkipVerify).
		WithTLSRootCAPool(c.ensureRootCAPool()).
		WithTransport(c.transport)
}

// NewClient returns a new vault client
func NewClient() *Client {
	return &Client{Scheme: SchemeTLS, transport: &http.Transport{}}
}

// NewClientFromEnv returns a new vault client with settings from environment
func NewClientFromEnv() (*Client, error) {
	if env.Env().Has(EnvVarVaultToken) && env.Env().Has(EnvVarVaultHost) {
		return &Client{
			Token:         env.Env().String(EnvVarVaultToken),
			Host:          env.Env().String(EnvVarVaultHost),
			Scheme:        SchemeTLS,
			TLSSkipVerify: env.Env().Bool(EnvVarVaultSkipVerify),
			CACertPath:    env.Env().String(EnvVarVaultCACert, kube.PathKubeCACert),
			transport:     &http.Transport{},
		}, nil
	}
	return nil, fmt.Errorf("vault: %s and %s required", EnvVarVaultHost, EnvVarVaultToken)
}

// WithToken sets the vault token for the client
func (c *Client) WithToken(token string) *Client {
	c.Token = token
	return c
}

// WithHost sets the vault host for the client
func (c *Client) WithHost(host string) *Client {
	c.Host = host
	return c
}

// WithScheme sets the vault host scheme
func (c *Client) WithScheme(scheme Scheme) *Client {
	c.Scheme = scheme
	return c
}

// WithTLS is a shortcut for WithScheme("https")
func (c *Client) WithTLS() *Client {
	return c.WithScheme(SchemeTLS)
}

// WithCACertPath sets the vault ca cert path
func (c *Client) WithCACertPath(path string) *Client {
	c.CACertPath = path
	return c
}

func (c *Client) ensureRootCAPool() *x509.CertPool {
	if c.rootCAPool == nil {
		certPool, err := x509.SystemCertPool()
		if err != nil {
			logging.Default().Error(exception.New(err))
			return nil
		}
		if len(c.CACertPath) == 0 {
			c.CACertPath = kube.PathKubeCACert
		}
		cert, err := ioutil.ReadFile(c.CACertPath)
		if os.IsNotExist(err) {
			logging.Default().Debugf("Kube ca cert does not exist at %s", kube.PathKubeCACert)
		} else if err != nil {
			logging.Default().Warning(exception.New(err))
		} else if !certPool.AppendCertsFromPEM(cert) {
			logging.Default().Warningf("No certificate added from %s", kube.PathKubeCACert)
		}
		c.rootCAPool = certPool
	}
	return c.rootCAPool
}

// HACK: define the function using var to make it mockable
var vaultRequest = func(req jsonRequest, successResponse interface{}) error {
	logEvent(req.Meta())
	var errorResponse vaultErrorResponse
	if successResponse == nil {
		successResponse = new(struct{})
	}

	tries := 0
	var responseMeta *request.ResponseMeta
	requestSucceeded := func() (bool, error) {
		tries++
		meta, err := req.JSONWithErrorHandler(successResponse, &errorResponse)
		if err != nil {
			if tries > maxRetries { // retries = tries - 1, maxRetries 0 = no retries
				return false, err
			}
			logging.Default().Warningf("Vault %s `%s` error: `%s`. retrying (%d)...", req.Meta().Method, req.Meta().URL, err.Error(), tries)
			return false, nil
		}
		responseMeta = meta
		return true, nil
	}
	if err := wait.PollImmediateInfinite(retryInterval, requestSucceeded); err != nil {
		return err
	}
	if req.Transport() != nil {
		req.Transport().CloseIdleConnections()
	}
	code := responseMeta.StatusCode
	if code == http.StatusOK || code == http.StatusNoContent {
		return nil
	}
	return &vaultError{
		code: code,
		error: fmt.Errorf("Vault %s `%s` error %d: %s", req.Meta().Method, req.Meta().URL, code,
			strings.Join(errorResponse.Errors, ", ")),
	}
}

func logEvent(meta *request.Meta) {
	if meta != nil && meta.URL != nil {
		url := meta.URL.Path
		verb := meta.Method
		logging.LogVaultRequest(url, verb)
	}
}
