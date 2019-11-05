package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	exception "github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/ref"
)

const (
	waitForDesiredCountSleep   = 10 * time.Second
	waitForDesiredCountTimeout = 20 * waitForDesiredCountSleep
	waitForDesiredCountError   = "CountsNotSame"
)

// DoesASGExist checks whether the given asg exists
func (a *AWS) DoesASGExist(name string) (bool, error) {
	g, err := a.GetASG(name)
	if err != nil {
		if ErrorCode(err) == ErrCodeNotFound {
			return false, nil
		}
		return false, exception.New(err)
	}
	return g != nil && g.AutoScalingGroupName != nil && *g.AutoScalingGroupName == name, nil
}

// GetASG returns the ASG with the specified name
func (a *AWS) GetASG(name string) (*autoscaling.Group, error) {
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{&name},
	}
	output, err := a.ensureASG().DescribeAutoScalingGroups(input)
	if err != nil {
		return nil, exception.New(err)
	}
	if len(output.AutoScalingGroups) == 0 {
		return nil, NewError(ErrCodeNotFound, "")
	}
	if len(output.AutoScalingGroups) != 1 {
		return nil, exception.New(fmt.Sprintf("Unexpected number of autoscaling groups found"))
	}
	return output.AutoScalingGroups[0], nil
}

// WaitForASGInService waits until the specified asg is in service
func (a *AWS) WaitForASGInService(name string) error {
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{&name},
	}
	return a.ensureASG().WaitUntilGroupInService(input)
}

// WaitForDesiredCount waits until the asg has the right number of instances
func (a *AWS) WaitForDesiredCount(name string, count int64) error {
	return a.DoAction(func() error {
		group, err := a.GetASG(name)
		if err != nil {
			return exception.New(err)
		}
		actual := int64(len(group.Instances))
		if actual == count {
			return nil
		}
		return NewError(waitForDesiredCountError, "")
	},
		func(err error) error {
			if err != nil {
				if ErrorCode(err) == waitForDesiredCountError {
					return nil
				}
				return exception.New(err)
			}
			return nil
		}, waitForDesiredCountSleep, waitForDesiredCountTimeout)
}

// CreateASG creates the asg
func (a *AWS) CreateASG(input *autoscaling.CreateAutoScalingGroupInput) error {
	_, err := a.ensureASG().CreateAutoScalingGroup(input)
	return exception.New(err)
}

// UpdateASG updates the specified autoscaling group
func (a *AWS) UpdateASG(input *autoscaling.UpdateAutoScalingGroupInput) error {
	_, err := a.ensureASG().UpdateAutoScalingGroup(input)
	return exception.New(err)
}

// UpdateASGTags updates the asg tags
func (a *AWS) UpdateASGTags(input *autoscaling.CreateOrUpdateTagsInput) error {
	_, err := a.ensureASG().CreateOrUpdateTags(input)
	return exception.New(err)
}

// DeleteASG deletes the specified asg
func (a *AWS) DeleteASG(name string) error {
	yes := true
	input := &autoscaling.DeleteAutoScalingGroupInput{
		AutoScalingGroupName: &name,
		ForceDelete:          &yes,
	}
	_, err := a.ensureASG().DeleteAutoScalingGroup(input)
	return exception.New(err)
}

// CycleASG cycles the autoscaling group
func (a *AWS) CycleASG(name string, capacity, maxCapacity int64) error {
	upscale := &autoscaling.UpdateAutoScalingGroupInput{
		DesiredCapacity:      &maxCapacity,
		AutoScalingGroupName: &name,
	}
	_, err := a.ensureASG().UpdateAutoScalingGroup(upscale)
	if err != nil {
		return exception.New(err)
	}
	err = a.WaitForASGInService(name)
	if err != nil {
		return exception.New(err)
	}
	err = a.WaitForDesiredCount(name, maxCapacity)
	if err != nil {
		return exception.New(err)
	}

	downscale := &autoscaling.UpdateAutoScalingGroupInput{
		DesiredCapacity:      &capacity,
		AutoScalingGroupName: &name,
	}
	_, err = a.ensureASG().UpdateAutoScalingGroup(downscale)
	if err != nil {
		return exception.New(err)
	}
	// Don't wait for downscale because it takes too long
	return a.WaitForASGInService(name)
}

// UpdateAndCycleASG updates the asg, waits, and then cycles it
func (a *AWS) UpdateAndCycleASG(input *autoscaling.UpdateAutoScalingGroupInput, tagsInput *autoscaling.CreateOrUpdateTagsInput) error {
	err := a.UpdateASG(input)
	if err != nil {
		return exception.New(err)
	}
	err = a.UpdateASGTags(tagsInput)
	if err != nil {
		return exception.New(err)
	}
	err = a.WaitForASGInService(*input.AutoScalingGroupName)
	if err != nil {
		return exception.New(err)
	}
	return a.CycleASG(*input.AutoScalingGroupName, *input.DesiredCapacity, *input.MaxSize)
}

// CreateLaunchConfigurationWithTimeout retires creating launch config until success of timeout
func (a *AWS) CreateLaunchConfigurationWithTimeout(input *autoscaling.CreateLaunchConfigurationInput, sleep time.Duration, timeout time.Duration) error {
	return a.DoAction(func() error {
		err := a.CreateLaunchConfiguration(input)
		return exception.New(err)
	}, IgnoreValidationErrors, sleep, timeout)
}

// CreateLaunchConfiguration creates the specified launch config
func (a *AWS) CreateLaunchConfiguration(input *autoscaling.CreateLaunchConfigurationInput) error {
	_, err := a.ensureASG().CreateLaunchConfiguration(input)
	return exception.New(err)
}

// GetLaunchConfiguration gets the specified launch configuration
func (a *AWS) GetLaunchConfiguration(name string) (*autoscaling.LaunchConfiguration, error) {
	input := &autoscaling.DescribeLaunchConfigurationsInput{
		LaunchConfigurationNames: []*string{&name},
	}
	output, err := a.ensureASG().DescribeLaunchConfigurations(input)
	if err != nil {
		return nil, exception.New(err)
	}
	if len(output.LaunchConfigurations) != 1 {
		return nil, exception.New(fmt.Sprintf("Unexpected number of launch configurations found"))
	}
	return output.LaunchConfigurations[0], nil
}

// AddLaunchConfigurationToASG adds the lc to the specified asg
func (a *AWS) AddLaunchConfigurationToASG(lcName, asgName string) error {
	input := &autoscaling.UpdateAutoScalingGroupInput{
		LaunchConfigurationName: &lcName,
		AutoScalingGroupName:    &asgName,
	}
	_, err := a.ensureASG().UpdateAutoScalingGroup(input)
	return exception.New(err)
}

// AddELBTargetGroupsToASG adds the target groups for the elb to the asg
func (a *AWS) AddELBTargetGroupsToASG(asgName string, elbName string) error {
	groups, err := a.ListTargetGroups(elbName)
	if err != nil {
		return exception.New(err)
	}
	arns := []*string{}
	for _, group := range groups {
		if group != nil {
			arns = append(arns, group.TargetGroupArn)
		}
	}
	input := &autoscaling.AttachLoadBalancerTargetGroupsInput{
		AutoScalingGroupName: &asgName,
		TargetGroupARNs:      arns,
	}
	_, err = a.ensureASG().AttachLoadBalancerTargetGroups(input)
	return exception.New(err)
}

// CreateAndUpdateLaunchConfiguration creates the lc and adds it to the asg
func (a *AWS) CreateAndUpdateLaunchConfiguration(input *autoscaling.CreateLaunchConfigurationInput, asgName string, desiredCapacity, maxCapacity int64) error {
	err := a.CreateLaunchConfiguration(input)
	if err != nil {
		return exception.New(err)
	}
	err = a.AddLaunchConfigurationToASG(*input.LaunchConfigurationName, asgName)
	if err != nil {
		return exception.New(err)
	}
	return a.CycleASG(asgName, desiredCapacity, maxCapacity)
}

// CreateAndUpdateLaunchConfigurationWithTimeout creates the lc and adds it to the asg
func (a *AWS) CreateAndUpdateLaunchConfigurationWithTimeout(input *autoscaling.CreateLaunchConfigurationInput, asgName string, desiredCapacity, maxCapacity int64, sleep, timeout time.Duration) error {
	err := a.CreateLaunchConfigurationWithTimeout(input, sleep, timeout)
	if err != nil {
		return exception.New(err)
	}
	err = a.AddLaunchConfigurationToASG(*input.LaunchConfigurationName, asgName)
	if err != nil {
		return exception.New(err)
	}
	return a.CycleASG(asgName, desiredCapacity, maxCapacity)
}

// ListLaunchConfigurations lists all launch configs
func (a *AWS) ListLaunchConfigurations() ([]*autoscaling.LaunchConfiguration, error) {
	oneHunna := int64(100)
	input := &autoscaling.DescribeLaunchConfigurationsInput{MaxRecords: &oneHunna}
	output, err := a.ensureASG().DescribeLaunchConfigurations(input)
	if err != nil {
		return nil, exception.New(err)
	}
	configs := output.LaunchConfigurations
	for output.NextToken != nil {
		input.NextToken = output.NextToken
		output, err = a.ensureASG().DescribeLaunchConfigurations(input)
		if err != nil {
			return nil, exception.New(err)
		}
		configs = append(configs, output.LaunchConfigurations...)
	}
	return configs, nil
}

// DeleteLaunchConfiguration deletes the specified launch configuration
func (a *AWS) DeleteLaunchConfiguration(name string) error {
	input := &autoscaling.DeleteLaunchConfigurationInput{
		LaunchConfigurationName: &name,
	}
	_, err := a.ensureASG().DeleteLaunchConfiguration(input)
	return exception.New(err)
}

// SetInstanceProtection sets the scale-in protection of the specified instances
func (a *AWS) SetInstanceProtection(name string, instanceIDs []string, protected bool) error {
	input := &autoscaling.SetInstanceProtectionInput{
		AutoScalingGroupName: &name,
		InstanceIds:          ref.Strings(instanceIDs...),
		ProtectedFromScaleIn: &protected,
	}
	_, err := a.ensureASG().SetInstanceProtection(input)
	return exception.New(err)
}

// TerminateInstanceInASG terminates an instance in an asg, optionally decrementing the desired capacity
func (a *AWS) TerminateInstanceInASG(instanceID string, shouldDecrement bool) error {
	input := &autoscaling.TerminateInstanceInAutoScalingGroupInput{
		InstanceId:                     &instanceID,
		ShouldDecrementDesiredCapacity: &shouldDecrement,
	}
	_, err := a.ensureASG().TerminateInstanceInAutoScalingGroup(input)
	return exception.New(err)
}
