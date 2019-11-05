package aws

import (
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	exception "github.com/blend/go-sdk/exception"
)

// ServiceMock is a mocked aws service
type ServiceMock struct {
	Responses chan interface{}
	Errors    chan error
}

func typeError(v interface{}) error {
	name := ""
	if v != nil {
		name = reflect.TypeOf(v).Name()
	}
	return exception.New(fmt.Sprintf("Invalid type for response `%s` `%v`", name, v))
}

// NewAWSServiceMock returns a new aws service mock
func NewAWSServiceMock(resps chan interface{}, errs chan error) ServiceMock {
	return ServiceMock{
		Responses: resps,
		Errors:    errs,
	}
}

// MockedAWS mocks aws
func MockedAWS() (*AWS, chan interface{}, chan error) {
	resps := make(chan interface{}, 20)
	errs := make(chan error, 20)
	serv := ServiceMock{
		Responses: resps,
		Errors:    errs,
	}
	return &AWS{
		ec2:     &EC2Mock{ServiceMock: serv},
		ecr:     &ECRMock{ServiceMock: serv},
		elb:     &ELBMock{ServiceMock: serv},
		elbv2:   &ELBv2Mock{ServiceMock: serv},
		route53: &Route53Mock{ServiceMock: serv},
		asg:     &ASGMock{ServiceMock: serv},
		iam:     &IAMMock{ServiceMock: serv},
		s3:      &S3Mock{ServiceMock: serv},
		cf:      &CFMock{ServiceMock: serv},
		kms:     &KMSMock{ServiceMock: serv},
		ses:     &SESMock{ServiceMock: serv},
	}, resps, errs
}

// EC2Mock mocks ec2
type EC2Mock struct {
	ec2iface.EC2API
	ServiceMock
}

// ECRMock mocks ecr
type ECRMock struct {
	ecriface.ECRAPI
	ServiceMock
}

// ELBMock mocks elb
type ELBMock struct {
	elbiface.ELBAPI
	ServiceMock
}

// ELBv2Mock mocks elbv2
type ELBv2Mock struct {
	elbv2iface.ELBV2API
	ServiceMock
}

// Route53Mock mocks route53
type Route53Mock struct {
	route53iface.Route53API
	ServiceMock
}

// ASGMock mocks autoscaling groups
type ASGMock struct {
	autoscalingiface.AutoScalingAPI
	ServiceMock
}

// IAMMock mocks iam
type IAMMock struct {
	iamiface.IAMAPI
	ServiceMock
}

// S3Mock mocks s3
type S3Mock struct {
	s3iface.S3API
	ServiceMock
}

// CFMock mocks cloudformation
type CFMock struct {
	cloudformationiface.CloudFormationAPI
	ServiceMock
}

// KMSMock mocks kms
type KMSMock struct {
	kmsiface.KMSAPI
	ServiceMock
}

// SESMock mocks ses
type SESMock struct {
	sesiface.SESAPI
	ServiceMock
}

//*******************************CloudFormation****************************

//DescribeStacks does stuff
func (m *CFMock) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*cloudformation.DescribeStacksOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

//CreateStack does stuff
func (m *CFMock) CreateStack(input *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*cloudformation.CreateStackOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

//UpdateStack does stuff
func (m *CFMock) UpdateStack(input *cloudformation.UpdateStackInput) (*cloudformation.UpdateStackOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*cloudformation.UpdateStackOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// WaitUntilStackCreateComplete does stuff
func (m *CFMock) WaitUntilStackCreateComplete(input *cloudformation.DescribeStacksInput) error {
	return <-m.Errors
}

// WaitUntilStackUpdateComplete does stuff
func (m *CFMock) WaitUntilStackUpdateComplete(input *cloudformation.DescribeStacksInput) error {
	return <-m.Errors
}

//*******************************KMS****************************

// CreateKey does stuff
func (m *KMSMock) CreateKey(input *kms.CreateKeyInput) (*kms.CreateKeyOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*kms.CreateKeyOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// CreateAlias does stuff
func (m *KMSMock) CreateAlias(input *kms.CreateAliasInput) (*kms.CreateAliasOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*kms.CreateAliasOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DescribeKey does stuff
func (m *KMSMock) DescribeKey(input *kms.DescribeKeyInput) (*kms.DescribeKeyOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*kms.DescribeKeyOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

//*******************************EC2****************************

// DescribeSecurityGroups does stuff
func (m *EC2Mock) DescribeSecurityGroups(input *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*ec2.DescribeSecurityGroupsOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DescribeVpcs does stuff
func (m *EC2Mock) DescribeVpcs(input *ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*ec2.DescribeVpcsOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DescribeSubnets does stuff
func (m *EC2Mock) DescribeSubnets(input *ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*ec2.DescribeSubnetsOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// CreateTags does stuff
func (m *EC2Mock) CreateTags(input *ec2.CreateTagsInput) (*ec2.CreateTagsOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*ec2.CreateTagsOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DeleteTags does stuff
func (m *EC2Mock) DeleteTags(input *ec2.DeleteTagsInput) (*ec2.DeleteTagsOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*ec2.DeleteTagsOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DeleteSecurityGroup does stuff
func (m *EC2Mock) DeleteSecurityGroup(input *ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*ec2.DeleteSecurityGroupOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// CreateSecurityGroup does stuff
func (m *EC2Mock) CreateSecurityGroup(input *ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*ec2.CreateSecurityGroupOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// RevokeSecurityGroupIngress does stuff
func (m *EC2Mock) RevokeSecurityGroupIngress(input *ec2.RevokeSecurityGroupIngressInput) (*ec2.RevokeSecurityGroupIngressOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*ec2.RevokeSecurityGroupIngressOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// RevokeSecurityGroupEgress does stuff
func (m *EC2Mock) RevokeSecurityGroupEgress(input *ec2.RevokeSecurityGroupEgressInput) (*ec2.RevokeSecurityGroupEgressOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*ec2.RevokeSecurityGroupEgressOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// AuthorizeSecurityGroupIngress does stuff
func (m *EC2Mock) AuthorizeSecurityGroupIngress(input *ec2.AuthorizeSecurityGroupIngressInput) (*ec2.AuthorizeSecurityGroupIngressOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*ec2.AuthorizeSecurityGroupIngressOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// AuthorizeSecurityGroupEgress does stuff
func (m *EC2Mock) AuthorizeSecurityGroupEgress(input *ec2.AuthorizeSecurityGroupEgressInput) (*ec2.AuthorizeSecurityGroupEgressOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*ec2.AuthorizeSecurityGroupEgressOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

//**************************************************************

//*******************************ECR****************************

// GetAuthorizationToken does stuff
func (m *ECRMock) GetAuthorizationToken(input *ecr.GetAuthorizationTokenInput) (*ecr.GetAuthorizationTokenOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*ecr.GetAuthorizationTokenOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

//**************************************************************

//******************************ELB***************************

// DescribeLoadBalancers describes elb loadbalancers
func (m *ELBMock) DescribeLoadBalancers(input *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*elb.DescribeLoadBalancersOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DescribeLoadBalancersPages describes elb loadbalancers w/ pagination
func (m *ELBMock) DescribeLoadBalancersPages(input *elb.DescribeLoadBalancersInput, fn func(*elb.DescribeLoadBalancersOutput, bool) bool) error {
	output, err := m.DescribeLoadBalancers(input)
	if err != nil {
		return err
	}
	for output != nil && len(aws.StringValue(output.NextMarker)) > 0 {
		if !fn(output, false) {
			return nil
		}
		output, err = m.DescribeLoadBalancers(input)
		if err != nil {
			return err
		}
	}
	fn(output, true)
	return nil
}

//**************************************************************

//******************************ELBV2***************************

// DescribeLoadBalancers does stuff
func (m *ELBv2Mock) DescribeLoadBalancers(input *elbv2.DescribeLoadBalancersInput) (*elbv2.DescribeLoadBalancersOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*elbv2.DescribeLoadBalancersOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DescribeTargetGroups does stuff
func (m *ELBv2Mock) DescribeTargetGroups(input *elbv2.DescribeTargetGroupsInput) (*elbv2.DescribeTargetGroupsOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*elbv2.DescribeTargetGroupsOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DescribeListeners does stuff
func (m *ELBv2Mock) DescribeListeners(input *elbv2.DescribeListenersInput) (*elbv2.DescribeListenersOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*elbv2.DescribeListenersOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// CreateLoadBalancer does stuff
func (m *ELBv2Mock) CreateLoadBalancer(input *elbv2.CreateLoadBalancerInput) (*elbv2.CreateLoadBalancerOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*elbv2.CreateLoadBalancerOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// CreateTargetGroup does stuff
func (m *ELBv2Mock) CreateTargetGroup(input *elbv2.CreateTargetGroupInput) (*elbv2.CreateTargetGroupOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*elbv2.CreateTargetGroupOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// CreateListener does stuff
func (m *ELBv2Mock) CreateListener(input *elbv2.CreateListenerInput) (*elbv2.CreateListenerOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*elbv2.CreateListenerOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// WaitUntilLoadBalancerAvailable does stuff
func (m *ELBv2Mock) WaitUntilLoadBalancerAvailable(input *elbv2.DescribeLoadBalancersInput) error {
	return <-m.Errors
}

// DeleteLoadBalancer does stuff
func (m *ELBv2Mock) DeleteLoadBalancer(input *elbv2.DeleteLoadBalancerInput) (*elbv2.DeleteLoadBalancerOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*elbv2.DeleteLoadBalancerOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DeleteTargetGroup does stuff
func (m *ELBv2Mock) DeleteTargetGroup(input *elbv2.DeleteTargetGroupInput) (*elbv2.DeleteTargetGroupOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*elbv2.DeleteTargetGroupOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

//**************************************************************

//*****************************Route53**************************

// ListHostedZones does stuff
func (m *Route53Mock) ListHostedZones(input *route53.ListHostedZonesInput) (*route53.ListHostedZonesOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*route53.ListHostedZonesOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// ListResourceRecordSets does stuff
func (m *Route53Mock) ListResourceRecordSets(input *route53.ListResourceRecordSetsInput) (*route53.ListResourceRecordSetsOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*route53.ListResourceRecordSetsOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// ChangeResourceRecordSets does stuff
func (m *Route53Mock) ChangeResourceRecordSets(input *route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*route53.ChangeResourceRecordSetsOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

//**************************************************************

//*******************************ASG****************************

// DescribeAutoScalingGroups does stuff
func (m *ASGMock) DescribeAutoScalingGroups(input *autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*autoscaling.DescribeAutoScalingGroupsOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// CreateAutoScalingGroup does stuff
func (m *ASGMock) CreateAutoScalingGroup(input *autoscaling.CreateAutoScalingGroupInput) (*autoscaling.CreateAutoScalingGroupOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*autoscaling.CreateAutoScalingGroupOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// UpdateAutoScalingGroup does stuff
func (m *ASGMock) UpdateAutoScalingGroup(input *autoscaling.UpdateAutoScalingGroupInput) (*autoscaling.UpdateAutoScalingGroupOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*autoscaling.UpdateAutoScalingGroupOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DeleteAutoScalingGroup does stuff
func (m *ASGMock) DeleteAutoScalingGroup(input *autoscaling.DeleteAutoScalingGroupInput) (*autoscaling.DeleteAutoScalingGroupOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*autoscaling.DeleteAutoScalingGroupOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// CreateOrUpdateTags does stuff
func (m *ASGMock) CreateOrUpdateTags(input *autoscaling.CreateOrUpdateTagsInput) (*autoscaling.CreateOrUpdateTagsOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*autoscaling.CreateOrUpdateTagsOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// WaitUntilGroupInService does stuff
func (m *ASGMock) WaitUntilGroupInService(input *autoscaling.DescribeAutoScalingGroupsInput) error {
	return <-m.Errors
}

// CreateLaunchConfiguration does stuff
func (m *ASGMock) CreateLaunchConfiguration(input *autoscaling.CreateLaunchConfigurationInput) (*autoscaling.CreateLaunchConfigurationOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*autoscaling.CreateLaunchConfigurationOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DeleteLaunchConfiguration does stuff
func (m *ASGMock) DeleteLaunchConfiguration(input *autoscaling.DeleteLaunchConfigurationInput) (*autoscaling.DeleteLaunchConfigurationOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*autoscaling.DeleteLaunchConfigurationOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DescribeLaunchConfigurations does stuff
func (m *ASGMock) DescribeLaunchConfigurations(input *autoscaling.DescribeLaunchConfigurationsInput) (*autoscaling.DescribeLaunchConfigurationsOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*autoscaling.DescribeLaunchConfigurationsOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// AttachLoadBalancerTargetGroups does stuff
func (m *ASGMock) AttachLoadBalancerTargetGroups(input *autoscaling.AttachLoadBalancerTargetGroupsInput) (*autoscaling.AttachLoadBalancerTargetGroupsOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*autoscaling.AttachLoadBalancerTargetGroupsOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// SetInstanceProtection does stuff
func (m *ASGMock) SetInstanceProtection(input *autoscaling.SetInstanceProtectionInput) (*autoscaling.SetInstanceProtectionOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*autoscaling.SetInstanceProtectionOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// TerminateInstanceInAutoScalingGroup does stuff
func (m *ASGMock) TerminateInstanceInAutoScalingGroup(input *autoscaling.TerminateInstanceInAutoScalingGroupInput) (*autoscaling.TerminateInstanceInAutoScalingGroupOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*autoscaling.TerminateInstanceInAutoScalingGroupOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

//**************************************************************

//*******************************IAM****************************

// GetRole gets the role
func (m *IAMMock) GetRole(input *iam.GetRoleInput) (*iam.GetRoleOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.GetRoleOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// CreateRole creates
func (m *IAMMock) CreateRole(input *iam.CreateRoleInput) (*iam.CreateRoleOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.CreateRoleOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// CreateInstanceProfile creates
func (m *IAMMock) CreateInstanceProfile(input *iam.CreateInstanceProfileInput) (*iam.CreateInstanceProfileOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.CreateInstanceProfileOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// AddRoleToInstanceProfile does stuff
func (m *IAMMock) AddRoleToInstanceProfile(input *iam.AddRoleToInstanceProfileInput) (*iam.AddRoleToInstanceProfileOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.AddRoleToInstanceProfileOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// PutRolePolicy does stuff
func (m *IAMMock) PutRolePolicy(input *iam.PutRolePolicyInput) (*iam.PutRolePolicyOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.PutRolePolicyOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// ListRolePoliciesPages does stuff
func (m *IAMMock) ListRolePoliciesPages(input *iam.ListRolePoliciesInput, consumer func(output *iam.ListRolePoliciesOutput, last bool) bool) error {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.ListRolePoliciesOutput); ok || response == nil {
		consumer(out, true)
		return err
	}
	panic(typeError(response))
}

// UpdateAssumeRolePolicy does stuff
func (m *IAMMock) UpdateAssumeRolePolicy(input *iam.UpdateAssumeRolePolicyInput) (*iam.UpdateAssumeRolePolicyOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.UpdateAssumeRolePolicyOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DeleteRolePolicy does stuff
func (m *IAMMock) DeleteRolePolicy(input *iam.DeleteRolePolicyInput) (*iam.DeleteRolePolicyOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.DeleteRolePolicyOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DeleteRole does stuff
func (m *IAMMock) DeleteRole(input *iam.DeleteRoleInput) (*iam.DeleteRoleOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.DeleteRoleOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DeleteInstanceProfile does stuff
func (m *IAMMock) DeleteInstanceProfile(input *iam.DeleteInstanceProfileInput) (*iam.DeleteInstanceProfileOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.DeleteInstanceProfileOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// RemoveRoleFromInstanceProfile does stuff
func (m *IAMMock) RemoveRoleFromInstanceProfile(input *iam.RemoveRoleFromInstanceProfileInput) (*iam.RemoveRoleFromInstanceProfileOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.RemoveRoleFromInstanceProfileOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// CreateUser does stuff
func (m *IAMMock) CreateUser(input *iam.CreateUserInput) (*iam.CreateUserOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.CreateUserOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// GetUser does stuff
func (m *IAMMock) GetUser(input *iam.GetUserInput) (*iam.GetUserOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.GetUserOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DeleteUser does stuff
func (m *IAMMock) DeleteUser(input *iam.DeleteUserInput) (*iam.DeleteUserOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.DeleteUserOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// ListAccessKeys does stuff
func (m *IAMMock) ListAccessKeys(input *iam.ListAccessKeysInput) (*iam.ListAccessKeysOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.ListAccessKeysOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// ListAccessKeysPages does stuff
func (m *IAMMock) ListAccessKeysPages(input *iam.ListAccessKeysInput, fn func(*iam.ListAccessKeysOutput, bool) bool) error {
	output, err := m.ListAccessKeys(input)
	if err != nil {
		return err
	}
	for output != nil && output.IsTruncated != nil && *output.IsTruncated {
		if !fn(output, false) {
			return nil
		}
		output, err = m.ListAccessKeys(input)
		if err != nil {
			return err
		}
	}
	fn(output, true)
	return nil
}

// ListUserPolicies does stuff
func (m *IAMMock) ListUserPolicies(input *iam.ListUserPoliciesInput) (*iam.ListUserPoliciesOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.ListUserPoliciesOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// ListUserPoliciesPages does stuff
func (m *IAMMock) ListUserPoliciesPages(input *iam.ListUserPoliciesInput, fn func(*iam.ListUserPoliciesOutput, bool) bool) error {
	output, err := m.ListUserPolicies(input)
	if err != nil {
		return err
	}
	for output != nil && output.IsTruncated != nil && *output.IsTruncated {
		if !fn(output, false) {
			return nil
		}
		output, err = m.ListUserPolicies(input)
		if err != nil {
			return err
		}
	}
	fn(output, true)
	return nil
}

// DeleteUserPolicy does stuff
func (m *IAMMock) DeleteUserPolicy(input *iam.DeleteUserPolicyInput) (*iam.DeleteUserPolicyOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.DeleteUserPolicyOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DeleteAccessKey does stuff
func (m *IAMMock) DeleteAccessKey(input *iam.DeleteAccessKeyInput) (*iam.DeleteAccessKeyOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.DeleteAccessKeyOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// CreateAccessKey does stuff
func (m *IAMMock) CreateAccessKey(input *iam.CreateAccessKeyInput) (*iam.CreateAccessKeyOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.CreateAccessKeyOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// PutUserPolicy does stuff
func (m *IAMMock) PutUserPolicy(input *iam.PutUserPolicyInput) (*iam.PutUserPolicyOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.PutUserPolicyOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// AttachRolePolicy does stuff
func (m *IAMMock) AttachRolePolicy(input *iam.AttachRolePolicyInput) (*iam.AttachRolePolicyOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.AttachRolePolicyOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// ListPoliciesPages does stuff
func (m *IAMMock) ListPoliciesPages(input *iam.ListPoliciesInput, consumer func(output *iam.ListPoliciesOutput, last bool) bool) error {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.ListPoliciesOutput); ok || response == nil {
		consumer(out, true)
		return err
	}
	panic(typeError(response))
}

// CreatePolicy does stuff
func (m *IAMMock) CreatePolicy(input *iam.CreatePolicyInput) (*iam.CreatePolicyOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.CreatePolicyOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DeletePolicy does stuff
func (m *IAMMock) DeletePolicy(input *iam.DeletePolicyInput) (*iam.DeletePolicyOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*iam.DeletePolicyOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

//**************************************************************

//*******************************S3*****************************

// ListObjects does stuff
func (m *S3Mock) ListObjects(input *s3.ListObjectsInput) (*s3.ListObjectsOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*s3.ListObjectsOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// ListObjectsPages does stuff
func (m *S3Mock) ListObjectsPages(input *s3.ListObjectsInput, consumer func(output *s3.ListObjectsOutput, last bool) bool) error {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*s3.ListObjectsOutput); ok || response == nil {
		consumer(out, true)
		return err
	}
	panic(typeError(response))
}

// DeleteObjects does stuff
func (m *S3Mock) DeleteObjects(input *s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*s3.DeleteObjectsOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DeleteObject does stuff
func (m *S3Mock) DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*s3.DeleteObjectOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// PutObject does stuff
func (m *S3Mock) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*s3.PutObjectOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// CreateBucket does stuff
func (m *S3Mock) CreateBucket(input *s3.CreateBucketInput) (*s3.CreateBucketOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*s3.CreateBucketOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// WaitUntilBucketExists does stuff
func (m *S3Mock) WaitUntilBucketExists(input *s3.HeadBucketInput) error {
	err := <-m.Errors
	return err
}

// PutBucketEncryption does stuff
func (m *S3Mock) PutBucketEncryption(input *s3.PutBucketEncryptionInput) (*s3.PutBucketEncryptionOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*s3.PutBucketEncryptionOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// DeleteBucket does stuff
func (m *S3Mock) DeleteBucket(input *s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*s3.DeleteBucketOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// GetBucketPolicy does stuff
func (m *S3Mock) GetBucketPolicy(input *s3.GetBucketPolicyInput) (*s3.GetBucketPolicyOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*s3.GetBucketPolicyOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// GetBucketLifecycleConfiguration does stuff
func (m *S3Mock) GetBucketLifecycleConfiguration(input *s3.GetBucketLifecycleConfigurationInput) (*s3.GetBucketLifecycleConfigurationOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*s3.GetBucketLifecycleConfigurationOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// PutBucketLifecycleConfiguration does stuff
func (m *S3Mock) PutBucketLifecycleConfiguration(input *s3.PutBucketLifecycleConfigurationInput) (*s3.PutBucketLifecycleConfigurationOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*s3.PutBucketLifecycleConfigurationOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// GetBucketReplication does stuff
func (m *S3Mock) GetBucketReplication(input *s3.GetBucketReplicationInput) (*s3.GetBucketReplicationOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*s3.GetBucketReplicationOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// PutBucketReplication does stuff
func (m *S3Mock) PutBucketReplication(input *s3.PutBucketReplicationInput) (*s3.PutBucketReplicationOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*s3.PutBucketReplicationOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

// PutBucketVersioning does stuff
func (m *S3Mock) PutBucketVersioning(input *s3.PutBucketVersioningInput) (*s3.PutBucketVersioningOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*s3.PutBucketVersioningOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

//**************************************************************

//*******************************SES****************************

// SendEmail does stuff
func (m *SESMock) SendEmail(input *ses.SendEmailInput) (*ses.SendEmailOutput, error) {
	response := <-m.Responses
	err := <-m.Errors
	if out, ok := response.(*ses.SendEmailOutput); ok || response == nil {
		return out, err
	}
	panic(typeError(response))
}

//**************************************************************
