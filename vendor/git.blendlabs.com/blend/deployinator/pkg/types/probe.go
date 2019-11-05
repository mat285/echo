package types

// Probe represents a liveness probe.
type Probe struct {
	Handler             `json:",inline" yaml:",inline" `
	InitialDelaySeconds int32 `json:"initialDelaySeconds,omitempty" yaml:"initialDelaySeconds,omitempty"`
	TimeoutSeconds      int32 `json:"timeoutSeconds,omitempty" yaml:"timeoutSeconds,omitempty"`
	PeriodSeconds       int32 `json:"periodSeconds,omitempty" yaml:"periodSeconds,omitempty"`
	SuccessThreshold    int32 `json:"successThreshold,omitempty" yaml:"successThreshold,omitempty"`
	FailureThreshold    int32 `json:"failureThreshold,omitempty" yaml:"failureThreshold,omitempty"`
}

// Handler defines a specific action that should be taken
type Handler struct {
	Exec      *ExecAction      `json:"exec,omitempty" yaml:"exec,omitempty"`
	HTTPGet   *HTTPGetAction   `json:"httpGet,omitempty" yaml:"httpGet,omitempty"`
	TCPSocket *TCPSocketAction `json:"tcpSocket,omitempty" yaml:"tcpSocket,omitempty"`
}

// ExecAction describes a "run in container" action.
type ExecAction struct {
	Command []string `json:"command,omitempty" yaml:"command,omitempty"`
}

// HTTPGetAction describes an action based on HTTP Get requests.
type HTTPGetAction struct {
	Path        string       `json:"path,omitempty" yaml:"path,omitempty"`
	Port        int          `json:"port,omitempty" yaml:"port,omitempty"`
	Host        string       `json:"host,omitempty" yaml:"host,omitempty"`
	Scheme      URIScheme    `json:"scheme,omitempty" string:"scheme,omitempty"`
	HTTPHeaders []HTTPHeader `json:"httpHeaders,omitempty" yaml:"httpHeaders,omitempty"`
}

// URIScheme identifies the scheme used for connection to a host for Get actions
type URIScheme string

const (
	// URISchemeHTTP means that the scheme used will be http://
	URISchemeHTTP URIScheme = "HTTP"
	// URISchemeHTTPS means that the scheme used will be https://
	URISchemeHTTPS URIScheme = "HTTPS"
)

// TCPSocketAction describes an action based on opening a socket
type TCPSocketAction struct {
	Port int    `json:"port" yaml:"port"`
	Host string `json:"host,omitempty" yaml:"host,omitempty"`
}

// HTTPHeader describes a custom header to be used in HTTP probes
type HTTPHeader struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}
