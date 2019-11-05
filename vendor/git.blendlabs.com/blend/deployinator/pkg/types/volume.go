package types

// Volume represents a named volume in a pod that may be accessed by any container in the pod.
type Volume struct {
	Name string `json:"name" yaml:"name"`

	HostPath             *HostPathVolumeSource             `json:"hostPath,omitempty" yaml:"hostPath,omitempty"`
	EmptyDir             *EmptyDirVolumeSource             `json:"emptyDir,omitempty" yaml:"emptyDir,omitempty"`
	AWSElasticBlockStore *AWSElasticBlockStoreVolumeSource `json:"awsElasticBlockStore,omitempty" yaml:"awsElasticBlockStore,omitempty"`
	GitRepo              *GitRepoVolumeSource              `json:"gitRepo,omitempty" yaml:"gitRepo,omitempty"`
	Secret               *SecretVolumeSource               `json:"secret,omitempty" yaml:"secret,omitempty"`
	ConfigMap            *ConfigMapVolumeSource            `json:"configMap,omitempty" yaml:"configMap,omitempty"`
	Projected            *ProjectedVolumeSource            `json:"projected,omitempty" yaml:"projected,omitempty"`
}

// HostPathVolumeSource represents a host path mapped into a pod.
type HostPathVolumeSource struct {
	Path string `json:"path" yaml:"path"`
}

// EmptyDirVolumeSource represents an empty directory for a pod.
type EmptyDirVolumeSource struct {
	Medium StorageMedium `json:"medium,omitempty" yaml:"medium,omitempty"`
}

// StorageMedium defines ways that storage can be allocated to a volume.
type StorageMedium string

const (
	// StorageMediumDefault is a constant.
	StorageMediumDefault StorageMedium = "" // use whatever the default is for the node
	// StorageMediumMemory is a constant.
	StorageMediumMemory StorageMedium = "Memory" // use memory (tmpfs)
)

// AWSElasticBlockStoreVolumeSource is a volume source
type AWSElasticBlockStoreVolumeSource struct {
	VolumeID  string `json:"volumeID" yaml:"volumeID"`
	FSType    string `json:"fsType,omitempty" yaml:"fsType,omitempty"`
	Partition int32  `json:"partition,omitempty" yaml:"partition,omitempty"`
	ReadOnly  bool   `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
}

// GitRepoVolumeSource represents a volume that is populated with the contents of a git repository.
// Git repo volumes do not support ownership management.
// Git repo volumes support SELinux relabeling.
type GitRepoVolumeSource struct {
	Repository string `json:"repository" yaml:"repository"`
	Revision   string `json:"revision,omitempty" yaml:"revision,omitempty"`
	Directory  string `json:"directory,omitempty" yaml:"directory,omitempty"`
}

// SecretVolumeSource adapts a Secret into a volume.
//
// The contents of the target Secret's Data field will be presented in a volume
// as files using the keys in the Data field as the file names.
// Secret volumes support ownership management and SELinux relabeling.
type SecretVolumeSource struct {
	SecretName  string      `json:"secretName,omitempty" yaml:"secretName,omitempty"`
	Items       []KeyToPath `json:"items,omitempty" yaml:"items,omitempty"`
	DefaultMode *int32      `json:"defaultMode,omitempty" yaml:"defaultMode,omitempty"`
	Optional    *bool       `json:"optional,omitempty" yaml:"optional,omitempty"`
}

// KeyToPath maps a string key to a path within a volume.
type KeyToPath struct {
	Key  string `json:"key" yaml:"key"`
	Path string `json:"path" yaml:"path"`
	Mode *int32 `json:"mode,omitempty" mode:"mode,omitempty"`
}

const (
	// SecretVolumeSourceDefaultMode is a default chflags permission set.
	SecretVolumeSourceDefaultMode int32 = 0644
)

// ConfigMapVolumeSource adapts a ConfigMap into a volume.
//
// The contents of the target ConfigMap's Data field will be presented in a
// volume as files using the keys in the Data field as the file names, unless
// the items element is populated with specific mappings of keys to paths.
// ConfigMap volumes support ownership management and SELinux relabeling.
type ConfigMapVolumeSource struct {
	Name        string      `json:"name,omitempty" yaml:"name,omitempty"`
	Items       []KeyToPath `json:"items,omitempty" yaml:"items,omitempty"`
	DefaultMode *int32      `json:"defaultMode,omitempty" yaml:"defaultMode,omitempty"`
	Optional    *bool       `json:"optional,omitempty" yaml:"optional,omitempty"`
}

// ProjectedVolumeSource represents a projected volume source
type ProjectedVolumeSource struct {
	Sources     []VolumeProjection `json:"sources" yaml:"sources"`
	DefaultMode *int32             `json:"defaultMode,omitempty" yaml:"defaultMode,omitempty"`
}

const (
	// ProjectedVolumeSourceDefaultMode is the default projected volume ch flags.
	ProjectedVolumeSourceDefaultMode int32 = 0644
)

// VolumeProjection is a projection that may be projected along with other supported volume types
type VolumeProjection struct {
	Secret      *SecretProjection      `json:"secret,omitempty" yaml:"secret,omitempty"`
	DownwardAPI *DownwardAPIProjection `json:"downwardAPI,omitempty" yaml:"downwardAPI,omitempty"`
	ConfigMap   *ConfigMapProjection   `json:"configMap,omitempty" yaml:"configMap,omitempty"`
}

// SecretProjection is a type.
type SecretProjection struct {
	Name     string      `json:"name,omitempty" yaml:"name,omitempty"`
	Items    []KeyToPath `json:"items,omitempty" yaml:"items,omitempty"`
	Optional *bool       `json:"optional,omitempty" yaml:"optional,omitempty"`
}

// ConfigMapProjection is a type.
type ConfigMapProjection struct {
	Name     string      `json:"name,omitempty" yaml:"name,omitempty"`
	Items    []KeyToPath `json:"items,omitempty" yaml:"items,omitempty"`
	Optional *bool       `json:"optional,omitempty" yaml:"optional,omitempty"`
}

// DownwardAPIProjection represents downward API info for projecting into a projected volume.
// Note that this is identical to a downwardAPI volume source without the default
// mode.
type DownwardAPIProjection struct {
	// Items is a list of DownwardAPIVolume files
	Items []DownwardAPIVolumeFile `json:"items,omitempty" protobuf:"bytes,1,rep,name=items"`
}

// DownwardAPIVolumeFile represents information to create the file containing the pod field
type DownwardAPIVolumeFile struct {
	// Required: Path is  the relative path name of the file to be created. Must not be absolute or contain the '..' path. Must be utf-8 encoded. The first item of the relative path must not start with '..'
	Path string `json:"path" yaml:"path"`
	// Required: Selects a field of the pod: only annotations, labels, name and namespace are supported.
	// +optional
	FieldRef *ObjectFieldSelector `json:"fieldRef,omitempty" yaml:"fieldRef,omitempty"`
	// Selects a resource of the container: only resources limits and requests
	// (limits.cpu, limits.memory, requests.cpu and requests.memory) are currently supported.
	// +optional
	ResourceFieldRef *ResourceFieldSelector `json:"resourceFieldRef,omitempty" yaml:"resourceFieldRef,omitempty"`
	// Optional: mode bits to use on this file, must be a value between 0
	// and 0777. If not specified, the volume defaultMode will be used.
	// This might be in conflict with other options that affect the file
	// mode, like fsGroup, and the result can be other mode bits set.
	// +optional
	Mode *int32 `json:"mode,omitempty" yaml:"mode,omitempty"`
}
