package types

// Protocol defines network protocols supported for things like container ports.
type Protocol string

const (
	// ProtocolTCP is the TCP protocol.
	ProtocolTCP Protocol = "TCP"
	// ProtocolUDP is the UDP protocol.
	ProtocolUDP Protocol = "UDP"
	// ProtocolHTTP is meant for use in service config
	ProtocolHTTP Protocol = "HTTP"
)

// ContainerPort represents a network port in a single container.
type ContainerPort struct {
	// If specified, this must be an IANA_SVC_NAME and unique within the pod. Each
	// named port in a pod must have a unique name. Name for the port that can be
	// referred to by services.
	// +optional
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Number of port to expose on the host.
	// If specified, this must be a valid port number, 0 < x < 65536.
	// If HostNetwork is specified, this must match ContainerPort.
	// Most containers do not need this.
	// +optional
	HostPort int32 `json:"hostPort,omitempty" yaml:"hostPort,omitempty"`
	// ContainerPort is the port exposed on the container that the service listens to
	// This must be a valid port number, 0 < x < 65536.
	ContainerPort int32 `json:"containerPort" yaml:"containerPort"`
	// ServicePort is the port exposed to external clients that want to access the server
	ServicePort int32 `json:"servicePort,omitempty" yaml:"servicePort,omitempty"`
	// Protocol for port. Must be UDP or TCP.
	// Defaults to "TCP".
	// +optional
	Protocol Protocol `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	// What host IP to bind the external port to.
	// +optional
	HostIP string `json:"hostIP,omitempty" yaml:"hostIP,omitempty"`
}

// EnsurePorts ensures that the service port is set
func (c *ContainerPort) EnsurePorts() {
	// for back compatability, if the service port is unset set it to the container port
	if c != nil && c.ServicePort <= 0 {
		c.ServicePort = c.ContainerPort
	}
}
