package types

import (
	"fmt"
	"time"

	"git.blendlabs.com/blend/deployinator/pkg/kube"
	"github.com/blend/go-sdk/exception"
	"k8s.io/api/core/v1"
)

// NewDeployFromPod returns a new deploy from a pod.
func NewDeployFromPod(pod *v1.Pod) Deploy {
	pinned := false
	if pod.Labels != nil {
		_, pinned = pod.Labels[kube.LabelPinPod]
	}
	status, _ := GetPodStatus(pod)
	deployedBy, _ := KubeUnescapeTarget(pod.ObjectMeta.Annotations[kube.AnnotationDeployedBy])
	deploy := Deploy{
		ID:           pod.Name,
		PodName:      pod.Name,
		Namespace:    pod.Namespace,
		ServiceName:  pod.Labels[kube.LabelService],
		ProjectName:  pod.Labels[kube.LabelProject],
		DatabaseName: pod.Labels[kube.LabelDatabase],
		DeployedAt:   pod.CreationTimestamp.UTC(),
		DeployedBy:   deployedBy,
		Phase:        string(pod.Status.Phase),
		BuildMode:    pod.ObjectMeta.Labels[kube.LabelBuildMode],
		Status:       status,
		IsPinned:     pinned,
	}
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodScheduled {
			deploy.ScheduledCondition = condition.DeepCopy()
		} else if condition.Type == v1.PodInitialized {
			deploy.InitializedCondition = condition.DeepCopy()
		}
	}
	return deploy
}

// GetPodStatus returns the status of the pod
func GetPodStatus(pod *v1.Pod) (*PodStatus, error) {
	var podProgress PodStatus
	containerStatuses := pod.Status.ContainerStatuses
	if len(containerStatuses) == 0 {
		return nil, exception.New(fmt.Errorf("No containers found"))
	}
	podProgress.PodName = pod.Name
	for _, containerStatus := range containerStatuses {
		containerState := containerStatus.State
		containerProgress := ContainerStatus{
			ContainerName: containerStatus.Name,
		}

		if containerState.Terminated != nil {
			containerProgress.ContainerComplete = true
			containerProgress.ExitCode = &containerState.Terminated.ExitCode
			containerProgress.ExitReason = &containerState.Terminated.Reason
		}
		podProgress.AddContainerStatus(containerProgress)
	}
	podProgress.SetPodCompleteStatus()
	return &podProgress, nil
}

// ContainerStatus is a status of a container in the deploy
type ContainerStatus struct {
	ContainerName     string  `json:"containerName"`
	ExitCode          *int32  `json:"exitCode,omitempty"`
	ExitReason        *string `json:"exitReason,omitempty"`
	ContainerComplete bool    `json:"containerComplete"`
}

// PodStatus is the status of the deploy pod
type PodStatus struct {
	PodComplete bool              `json:"podComplete"`
	PodName     string            `json:"podName"`
	Containers  []ContainerStatus `json:"containers"`
}

// Deploy represents a build.
type Deploy struct {
	ID                   string           `json:"id" yaml:"id"`
	Namespace            string           `json:"namespace" yaml:"namespace"`
	ServiceName          string           `json:"serviceName" yaml:"serviceName"`
	ProjectName          string           `json:"projectName" yaml:"projectName"`
	DatabaseName         string           `json:"databaseName" yaml:"databaseName"`
	PodName              string           `json:"podName" yaml:"podName"`
	DeployedAt           time.Time        `json:"deployedAt" yaml:"deployedAt"`
	DeployedBy           string           `json:"deployedBy" yaml:"deployedBy"`
	Phase                string           `json:"phase" yaml:"phase"`
	ScheduledCondition   *v1.PodCondition `json:"scheduledCondition" yaml:"scheduledCondition"`
	InitializedCondition *v1.PodCondition `json:"initializedCondition" yaml:"initializedCondition"`
	BuildMode            string           `json:"buildMode" yaml:"buildMode"`
	Config               *Config          `json:"config" yaml:"config"`
	Status               *PodStatus       `json:"status" yaml:"status"`
	IsArchived           bool             `json:"isArchived" yaml:"isArchived"`
	IsPinned             bool             `json:"isPinned,omitempty" yaml:"isArchived,omitempty"`
}

// Deploys is a list of deploys.
type Deploys []Deploy

// Len returns the number of elements.
func (d Deploys) Len() int {
	return len(d)
}

// Swap swaps elements.
func (d Deploys) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

// Less returns if two elements are less than eachother.
func (d Deploys) Less(i, j int) bool {
	return d[i].DeployedAt.After(d[j].DeployedAt)
}

// ScopeOfDeploy returns the scope of the deploy
func ScopeOfDeploy(deploy *Deploy) (string, error) {
	if deploy == nil {
		return "", exception.New(fmt.Sprintf("Nil deploy"))
	}
	if len(deploy.ServiceName) > 0 {
		return ServiceScope(deploy.ServiceName), nil
	}
	if len(deploy.ProjectName) > 0 {
		return ProjectScope(deploy.ProjectName), nil
	}
	if len(deploy.DatabaseName) > 0 {
		return DatabaseScope(deploy.DatabaseName), nil
	}
	return "", exception.New(fmt.Sprintf("Missing project, service, and database name"))
}

// AddContainerStatus adds container statuses to the PodStatus
func (pp *PodStatus) AddContainerStatus(containerStatus ...ContainerStatus) {
	pp.Containers = append(pp.Containers, containerStatus...)
}

// SetPodCompleteStatus sets the pod completion status from the containers. Use this after adding all containers
func (pp *PodStatus) SetPodCompleteStatus() {
	complete := true
	for _, container := range pp.Containers {
		if !container.ContainerComplete {
			complete = false
		}
	}
	pp.PodComplete = complete
}
