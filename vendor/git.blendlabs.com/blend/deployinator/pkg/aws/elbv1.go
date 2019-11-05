package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/blend/go-sdk/collections"
	exception "github.com/blend/go-sdk/exception"
)

// GetELBsV1ByDNSName returns the elbs by dns name
func (a *AWS) GetELBsV1ByDNSName(names ...string) ([]*elb.LoadBalancerDescription, error) {
	elbs, err := a.GetELBsV1()
	if err != nil {
		return nil, err
	}
	set := collections.NewSetOfString(names...)

	ret := []*elb.LoadBalancerDescription{}
	for _, lb := range elbs {
		if lb != nil && set.Contains(aws.StringValue(lb.DNSName)) {
			ret = append(ret, lb)
		}
	}
	return ret, nil
}

// GetELBV1ByDNSName returns the elb if it exists. If not it returns nil and a notfound exception
func (a *AWS) GetELBV1ByDNSName(dnsName string) (*elb.LoadBalancerDescription, error) {
	elbs, err := a.GetELBsV1ByDNSName(dnsName)
	if err != nil {
		return nil, err
	}
	if len(elbs) != 1 || elbs[0] == nil {
		return nil, NewError(elb.ErrCodeAccessPointNotFoundException, "", nil)
	}
	return elbs[0], nil
}

// GetELBsV1 gets all of the elbs v1 (yes its the correct plural)
func (a *AWS) GetELBsV1() ([]*elb.LoadBalancerDescription, error) {
	input := &elb.DescribeLoadBalancersInput{}
	elbs := []*elb.LoadBalancerDescription{}
	err := a.ensureELBV1().DescribeLoadBalancersPages(input, func(output *elb.DescribeLoadBalancersOutput, _ bool) bool {
		elbs = append(elbs, output.LoadBalancerDescriptions...)
		return true
	})
	if err != nil {
		return nil, exception.New(err)
	}
	return elbs, nil
}

// ApplySecurityGroupsToLoadBalancer applies the security groups to the elb
func (a *AWS) ApplySecurityGroupsToLoadBalancer(lbName string, securityGroupIds ...*string) error {
	input := &elb.ApplySecurityGroupsToLoadBalancerInput{
		LoadBalancerName: &lbName,
		SecurityGroups:   securityGroupIds,
	}
	_, err := a.ensureELBV1().ApplySecurityGroupsToLoadBalancer(input)
	return exception.New(err)
}
