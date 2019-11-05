package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	exception "github.com/blend/go-sdk/exception"
)

const (
	waitForELBReadySleep   = 1 * time.Second
	waitForELBReadyTimeout = 20 * time.Minute
)

// DoesELBExist tests whether an elb with the given name exists
func (a *AWS) DoesELBExist(name string) (bool, error) {
	lb, err := a.GetELB(name)
	if err != nil {
		notFound := IgnoreErrorCodes(err, elbv2.ErrCodeLoadBalancerNotFoundException)
		return false, notFound
	}
	return lb != nil && lb.LoadBalancerName != nil && *lb.LoadBalancerName == name, nil
}

// GetELB returns the elb if it exists. If not it returns nil and a notfound exception
func (a *AWS) GetELB(name string) (*elbv2.LoadBalancer, error) {
	input := &elbv2.DescribeLoadBalancersInput{
		Names: []*string{aws.String(name)},
	}

	output, err := a.ensureELBV2().DescribeLoadBalancers(input)
	if err != nil {
		return nil, exception.New(err)
	}
	for _, lb := range output.LoadBalancers {
		if lb != nil && lb.LoadBalancerName != nil && *lb.LoadBalancerName == name {
			return lb, nil
		}
	}
	return nil, NewError(elbv2.ErrCodeLoadBalancerNotFoundException, "", nil)
}

// GetTG returns a target group if it exists. If not it returns nil and a notfound exception
func (a *AWS) GetTG(name string) (*elbv2.TargetGroup, error) {
	input := &elbv2.DescribeTargetGroupsInput{
		Names: []*string{aws.String(name)},
	}

	output, err := a.ensureELBV2().DescribeTargetGroups(input)
	if err != nil {
		return nil, exception.New(err)
	}

	for _, tg := range output.TargetGroups {
		if tg != nil && tg.TargetGroupName != nil && *tg.TargetGroupName == name {
			return tg, nil
		}
	}
	return nil, NewError(elbv2.ErrCodeTargetGroupNotFoundException, "", nil)
}

// DoesListenerExist returns whether a listener on the specified port exists on the elb
func (a *AWS) DoesListenerExist(elbARN string, port int64) (bool, error) {
	listeners, err := a.ListELBListeners(elbARN)
	if err != nil {
		return false, exception.New(err)
	}
	for _, listener := range listeners {
		if listener != nil && listener.Port != nil && *listener.Port == port {
			return true, nil
		}
	}
	return false, nil
}

// ListELBListeners lists all of the listeners on the specified elb
func (a *AWS) ListELBListeners(elbArn string) ([]*elbv2.Listener, error) {
	input := &elbv2.DescribeListenersInput{
		LoadBalancerArn: &elbArn,
	}
	output, err := a.ensureELBV2().DescribeListeners(input)
	if err != nil {
		return nil, exception.New(err)
	}
	listeners := output.Listeners
	for output.NextMarker != nil {
		input.Marker = output.NextMarker
		output, err = a.ensureELBV2().DescribeListeners(input)
		if err != nil {
			return nil, exception.New(err)
		}
		listeners = append(listeners, output.Listeners...)
	}
	return listeners, nil
}

// ListTargetGroups lists the target groups for the elb
func (a *AWS) ListTargetGroups(elbName string) ([]*elbv2.TargetGroup, error) {
	e, err := a.GetELB(elbName)
	if err != nil {
		return nil, exception.New(err)
	}
	input := &elbv2.DescribeTargetGroupsInput{
		LoadBalancerArn: e.LoadBalancerArn,
	}
	output, err := a.ensureELBV2().DescribeTargetGroups(input)
	if err != nil {
		return nil, exception.New(err)
	}
	groups := output.TargetGroups
	for output.NextMarker != nil {
		input.Marker = output.NextMarker
		output, err = a.ensureELBV2().DescribeTargetGroups(input)
		if err != nil {
			return nil, exception.New(err)
		}
		groups = append(groups, output.TargetGroups...)
	}
	return groups, nil
}

// CreateELBWithListener creates an elb and a listener on that elb
func (a *AWS) CreateELBWithListener(input *elbv2.CreateLoadBalancerInput, vpcID string, port int64, protocol string) (*elbv2.LoadBalancer, error) {
	elb, err := a.CreateELB(input)
	if err != nil {
		return nil, exception.New(err)
	}
	if elb == nil || elb.LoadBalancerArn == nil {
		return nil, exception.New(fmt.Sprintf("No load balancer created"))
	}
	return elb, a.CreateListener(port, protocol, *elb.LoadBalancerArn, *input.Name, vpcID)
}

// CreateTargetGroup creates a target group
func (a *AWS) CreateTargetGroup(elbName string, vpcID string, port int64, protocol string) (string, error) {
	ten := int64(10)
	targetType := elbv2.TargetTypeEnumInstance
	input := &elbv2.CreateTargetGroupInput{
		Name:                       aws.String(elbName),
		HealthCheckIntervalSeconds: &ten,
		TargetType:                 &targetType,
		VpcId:                      aws.String(vpcID),
		Port:                       aws.Int64(port),
		Protocol:                   aws.String(protocol),
	}

	output, err := a.ensureELBV2().CreateTargetGroup(input)
	if err != nil {
		return "", err
	}

	if output == nil {
		return "", exception.New(fmt.Sprintf("CreateTargetGroup output is nil"))
	}
	if len(output.TargetGroups) != 1 {
		return "", exception.New(fmt.Sprintf("CreateTargetGroup invalid number of target groups %d", len(output.TargetGroups)))
	}
	if output.TargetGroups[0].TargetGroupArn == nil {
		return "", exception.New(fmt.Sprintf("CreateTargetGroup arn is nil"))
	}

	return *output.TargetGroups[0].TargetGroupArn, nil
}

// CreateListener creates a listener with the specified port and protocol on the loadbalancer
func (a *AWS) CreateListener(port int64, protocol string, elbArn string, elbName string, vpcID string) error {
	targetGroupArn, err := a.CreateTargetGroup(elbName, vpcID, port, protocol)
	if err != nil {
		return exception.New(err)
	}

	action := elbv2.ActionTypeEnumForward
	input := &elbv2.CreateListenerInput{
		Port:            &port,
		Protocol:        aws.String(protocol),
		LoadBalancerArn: aws.String(elbArn),
		DefaultActions: []*elbv2.Action{
			&elbv2.Action{
				TargetGroupArn: aws.String(targetGroupArn),
				Type:           &action,
			},
		},
	}
	_, err = a.ensureELBV2().CreateListener(input)
	return exception.New(err)
}

// CreateELB creates a new ELB
func (a *AWS) CreateELB(input *elbv2.CreateLoadBalancerInput) (*elbv2.LoadBalancer, error) {
	output, err := a.ensureELBV2().CreateLoadBalancer(input)
	if err != nil {
		return nil, exception.New(err)
	}
	if len(output.LoadBalancers) != 1 {
		return nil, exception.New(fmt.Sprintf("Unexpected number of load balancers returned"))
	}
	lb := output.LoadBalancers[0]

	err = a.AwaitELBReady(input.Name)
	if err != nil {
		return nil, exception.New(err)
	}
	return lb, nil
}

// AwaitELBReady waits until elb is ready for use
func (a *AWS) AwaitELBReady(name *string) error {
	// This can take a really long time, wait for it twice if need be
	return a.DoAction(func() error {
		input := &elbv2.DescribeLoadBalancersInput{
			Names: []*string{name},
		}
		return a.ensureELBV2().WaitUntilLoadBalancerAvailable(input)
	}, func(err error) error {
		return nil
	}, waitForELBReadySleep, waitForELBReadyTimeout)
}

// DeleteELB removes the specified elbv2.
func (a *AWS) DeleteELB(name string) error {
	lb, err := a.GetELB(name)
	if err != nil {
		return exception.New(err)
	}
	input := &elbv2.DeleteLoadBalancerInput{
		LoadBalancerArn: lb.LoadBalancerArn,
	}
	_, err = a.ensureELBV2().DeleteLoadBalancer(input)
	if err != nil {
		return exception.New(err)
	}

	return a.DeleteTargetGroup(name)
}

// DeleteTargetGroup removes the specified target group.
func (a *AWS) DeleteTargetGroup(name string) error {
	tg, err := a.GetTG(name)
	if err != nil {
		return exception.New(err)
	}

	input := &elbv2.DeleteTargetGroupInput{
		TargetGroupArn: tg.TargetGroupArn,
	}

	_, err = a.ensureELBV2().DeleteTargetGroup(input)
	return exception.New(err)
}
