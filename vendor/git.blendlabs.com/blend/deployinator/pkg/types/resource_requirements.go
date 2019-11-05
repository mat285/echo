package types

// ResourceRequirements describes the compute resource requirements.
type ResourceRequirements struct {
	Limits   ResourceList `json:"limits,omitempty" yaml:"limits,omitempty"`
	Requests ResourceList `json:"requests,omitempty" yaml:"requests,omitempty"`
}

// ResourceList is a set of (resource name, quantity) pairs.
type ResourceList map[ResourceName]Quantity

// ResourceName is the name identifying various resources in a ResourceList.
type ResourceName string

const (
	// ResourceCPU is CPU, in cores. (500m = .5 cores)
	ResourceCPU ResourceName = "cpu"
	// ResourceMemory is Memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	ResourceMemory ResourceName = "memory"
	// ResourceStorage is Volume size, in bytes (e,g. 5Gi = 5GiB = 5 * 1024 * 1024 * 1024)
	ResourceStorage ResourceName = "storage"
)

// Quantity is just a string in our schema, we don't care about the internal implications.
type Quantity string
