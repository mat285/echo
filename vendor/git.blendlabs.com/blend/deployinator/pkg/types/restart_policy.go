package types

// RestartPolicy is a restart policy for the deployment / replica set.
type RestartPolicy string

const (
	// RestartPolicyAlways is a restart policy.
	RestartPolicyAlways RestartPolicy = "Always"
	// RestartPolicyOnFailure is a restart policy.
	RestartPolicyOnFailure RestartPolicy = "OnFailure"
	// RestartPolicyNever is a restart policy.
	RestartPolicyNever RestartPolicy = "Never"
)
