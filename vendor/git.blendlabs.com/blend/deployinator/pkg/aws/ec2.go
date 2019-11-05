package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	exception "github.com/blend/go-sdk/exception"
)

// GetSecurityGroupByName returns the security group by its name, errors if more than one is returned
func (a *AWS) GetSecurityGroupByName(name string) (*ec2.SecurityGroup, error) {
	sgs, err := a.GetSecurityGroupsByName(&name)
	if err != nil {
		return nil, err
	}
	if len(sgs) == 0 {
		return nil, NewError(ErrCodeNotFound, "Security Group Not Found")
	}
	if len(sgs) != 1 || sgs[0] == nil || sgs[0].GroupId == nil {
		return nil, exception.New("Unexpected Output From AWS")
	}
	return sgs[0], nil
}

// GetSecurityGroupIDsByName gets the aws security group identifiers by their names
func (a *AWS) GetSecurityGroupIDsByName(securityGroupNames []*string) ([]*string, error) {
	var securityGroupIds = []*string{}

	secGroups, err := a.GetSecurityGroupsByName(securityGroupNames...)
	if err != nil {
		return nil, exception.New(err)
	}
	for _, group := range secGroups {
		if group != nil && group.GroupId != nil {
			securityGroupIds = append(securityGroupIds, group.GroupId)
		}
	}
	return securityGroupIds, nil
}

// GetVPC gets the vpc for the id
func (a *AWS) GetVPC(id string) (*ec2.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		VpcIds: []*string{&id},
	}
	output, err := a.ensureEC2().DescribeVpcs(input)
	if err != nil {
		return nil, exception.New(err)
	}
	if output == nil || len(output.Vpcs) != 1 {
		return nil, exception.New(fmt.Sprintf("Unexpected number of VPCs"))
	}
	return output.Vpcs[0], nil
}

// listAllSubnetsForVPCFiltered lists the subnets in this vpc with optional filters
func (a *AWS) listAllSubnetsForVPCFiltered(vpcID string, filters ...*ec2.Filter) ([]*ec2.Subnet, error) {
	inputFilters := []*ec2.Filter{
		{
			Name: aws.String("vpc-id"),
			Values: []*string{
				aws.String(vpcID),
			},
		},
	}
	if len(filters) > 0 {
		inputFilters = append(inputFilters, filters...)
	}

	input := &ec2.DescribeSubnetsInput{
		Filters: inputFilters,
	}
	output, err := a.ensureEC2().DescribeSubnets(input)
	if err != nil {
		return nil, exception.New(err)
	}
	return output.Subnets, nil
}

// ListAllSubnetsForVPC lists the subnets in this vpc
func (a *AWS) ListAllSubnetsForVPC(vpcID string) ([]*ec2.Subnet, error) {
	return a.listAllSubnetsForVPCFiltered(vpcID)
}

// ListPublicSubnetsForVPC lists the public subnets in a vpc
func (a *AWS) ListPublicSubnetsForVPC(vpcID string) ([]*ec2.Subnet, error) {
	filter := &ec2.Filter{
		Name:   aws.String(fmt.Sprintf("tag:%s", TagSubnetAccessibilityTypeKey)),
		Values: []*string{aws.String(SubnetTypePublic)},
	}
	return a.listAllSubnetsForVPCFiltered(vpcID, filter)
}

// ListPrivateSubnetsForVPC lists the public subnets in a vpc
func (a *AWS) ListPrivateSubnetsForVPC(vpcID string) ([]*ec2.Subnet, error) {
	filter := &ec2.Filter{
		Name:   aws.String(fmt.Sprintf("tag:%s", TagSubnetAccessibilityTypeKey)),
		Values: []*string{aws.String(SubnetTypePrivate)},
	}
	return a.listAllSubnetsForVPCFiltered(vpcID, filter)
}

// UpsertTagsForEC2Resources adds tags to ec2 resources
func (a *AWS) UpsertTagsForEC2Resources(resourceIds []*string, tags []*ec2.Tag) error {
	input := &ec2.CreateTagsInput{
		Tags:      tags,
		Resources: resourceIds,
	}
	_, err := a.ensureEC2().CreateTags(input)
	return exception.New(err)
}

// DeleteTagsForEC2Resources deletes tags from ec2 resources
func (a *AWS) DeleteTagsForEC2Resources(resourceIds []*string, tags []*ec2.Tag) error {
	input := &ec2.DeleteTagsInput{
		Tags:      tags,
		Resources: resourceIds,
	}
	_, err := a.ensureEC2().DeleteTags(input)
	return exception.New(err)
}

// GetSecurityGroupsByName gets security groups by name
func (a *AWS) GetSecurityGroupsByName(names ...*string) ([]*ec2.SecurityGroup, error) {
	if len(names) == 0 {
		return []*ec2.SecurityGroup{}, nil
	}
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("group-name"),
				Values: names,
			},
		},
	}
	output, err := a.ensureEC2().DescribeSecurityGroups(input)
	if err != nil {
		return nil, exception.New(err)
	}
	return output.SecurityGroups, nil
}

// RemoveSecGroupPermissions removes a security group's permissions ingress and egress
func (a *AWS) RemoveSecGroupPermissions(secGroup *ec2.SecurityGroup) error {
	if secGroup == nil {
		return exception.New(fmt.Sprintf("nil security group input"))
	}
	if len(secGroup.IpPermissions) > 0 {
		input := &ec2.RevokeSecurityGroupIngressInput{
			GroupId:       secGroup.GroupId,
			IpPermissions: secGroup.IpPermissions,
		}
		_, err := a.ensureEC2().RevokeSecurityGroupIngress(input)
		if err != nil {
			return exception.New(err)
		}
	}

	if len(secGroup.IpPermissionsEgress) > 0 {
		inputEgress := &ec2.RevokeSecurityGroupEgressInput{
			GroupId:       secGroup.GroupId,
			IpPermissions: secGroup.IpPermissionsEgress,
		}
		_, err := a.ensureEC2().RevokeSecurityGroupEgress(inputEgress)
		if err != nil {
			return exception.New(err)
		}
	}
	return nil
}

// DeleteSecGroup removes a security group and its permissions
func (a *AWS) DeleteSecGroup(secGroup *ec2.SecurityGroup) error {
	if secGroup == nil {
		return nil
	}
	err := a.RemoveSecGroupPermissions(secGroup)
	if err != nil {
		return exception.New(err)
	}
	input := &ec2.DeleteSecurityGroupInput{
		GroupId: secGroup.GroupId,
	}
	_, err = a.ensureEC2().DeleteSecurityGroup(input)

	return exception.New(err)
}

// RemoveSecGroups removes security groups and their permissions
func (a *AWS) RemoveSecGroups(secGroups ...*ec2.SecurityGroup) error {
	errs := []string{}
	for _, secGroup := range secGroups {
		err := a.DeleteSecGroup(secGroup)
		if err != nil {
			errs = append(errs, fmt.Sprintf("* %s", err))
		}
	}
	if len(errs) != 0 {
		return exception.New(fmt.Sprintf("Found %d errors:\n%s", len(errs), strings.Join(errs, "\n")))
	}
	return nil
}

// CreateSecurityGroupIfNotExist creates and returns the security group if it doesn't exist
func (a *AWS) CreateSecurityGroupIfNotExist(name, vcpID, description string) (*ec2.SecurityGroup, error) {
	sgs, err := a.GetSecurityGroupsByName(&name)
	if err != nil {
		return nil, err
	}
	if len(sgs) > 1 {
		return nil, exception.New("TooManySecurityGroup")
	}
	if len(sgs) == 1 {
		return sgs[0], nil
	}
	id, err := a.CreateSecurityGroup(name, vcpID, description)
	if err != nil {
		return nil, err
	}
	return a.GetSecurityGroup(id)
}

// CreateSecurityGroup creates a security group
func (a *AWS) CreateSecurityGroup(name string, vpcID string, description string) (string, error) {
	input := &ec2.CreateSecurityGroupInput{
		Description: &description,
		GroupName:   &name,
		VpcId:       &vpcID,
	}
	output, err := a.ensureEC2().CreateSecurityGroup(input)
	if err != nil {
		return "", err
	}
	if output == nil {
		return "", exception.New(fmt.Sprintf("nil output returned from amazon"))
	}
	if output.GroupId == nil {
		return "", exception.New(fmt.Sprintf("nil groupId returned from amazon"))
	}
	return *output.GroupId, nil
}

// AddIngressRuleForSecGroup adds a rule to group from sourceGroup
func (a *AWS) AddIngressRuleForSecGroup(groupID, vpcID, sourceGroupID, protocol string) error {
	return a.AuthorizeSecurityGroupIngress(groupID, &ec2.IpPermission{
		IpProtocol: &protocol,
		FromPort:   aws.Int64(0),
		ToPort:     aws.Int64(65535),
		UserIdGroupPairs: []*ec2.UserIdGroupPair{
			&ec2.UserIdGroupPair{
				GroupId: &sourceGroupID,
				VpcId:   &vpcID,
			},
		},
	})
}

// AuthorizeSecurityGroupIngress authorizes security group ingress
func (a *AWS) AuthorizeSecurityGroupIngress(groupID string, rules ...*ec2.IpPermission) error {
	input := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:       &groupID,
		IpPermissions: rules,
	}
	_, err := a.ensureEC2().AuthorizeSecurityGroupIngress(input)
	err = IgnoreErrorCodes(err, ErrCodeDuplicateIPPermission)
	return exception.New(err)
}

// RemoveIngressRuleFromSecGroup removes the rule from the security group
func (a *AWS) RemoveIngressRuleFromSecGroup(groupID, vpcID, sourceGroupID, protocol string) error {
	input := &ec2.RevokeSecurityGroupIngressInput{
		GroupId: &groupID,
		IpPermissions: []*ec2.IpPermission{
			&ec2.IpPermission{
				IpProtocol: &protocol,
				FromPort:   aws.Int64(0),
				ToPort:     aws.Int64(65535),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					&ec2.UserIdGroupPair{
						GroupId: &sourceGroupID,
						VpcId:   &vpcID,
					},
				},
			},
		},
	}
	_, err := a.ensureEC2().RevokeSecurityGroupIngress(input)
	err = IgnoreErrorCodes(err, ErrCodeIPPermissionNotFound)
	return exception.New(err)
}

// AddIngressRuleToSecGroup adds ingress rule to a security group
func (a *AWS) AddIngressRuleToSecGroup(groupID string, cidrIP string, protocol string, fromPort int64, toPort int64) error {
	ipRange := &ec2.IpRange{
		CidrIp: &cidrIP,
	}
	ipPermission := &ec2.IpPermission{
		IpProtocol: &protocol,
		FromPort:   &fromPort,
		ToPort:     &toPort,
		IpRanges:   []*ec2.IpRange{ipRange},
	}
	input := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:       &groupID,
		IpPermissions: []*ec2.IpPermission{ipPermission},
	}
	_, err := a.ensureEC2().AuthorizeSecurityGroupIngress(input)
	return exception.New(err)
}

// AddEgressRuleToSecGroup adds egress rule to a security group
func (a *AWS) AddEgressRuleToSecGroup(groupID string, cidrIP string, protocol string, fromPort int64, toPort int64) error {
	ipRange := &ec2.IpRange{
		CidrIp: &cidrIP,
	}
	ipPermission := &ec2.IpPermission{
		IpProtocol: &protocol,
		FromPort:   &fromPort,
		ToPort:     &toPort,
		IpRanges:   []*ec2.IpRange{ipRange},
	}
	input := &ec2.AuthorizeSecurityGroupEgressInput{
		GroupId:       &groupID,
		IpPermissions: []*ec2.IpPermission{ipPermission},
	}
	_, err := a.ensureEC2().AuthorizeSecurityGroupEgress(input)
	return exception.New(err)
}

// GetSecurityGroup gets the security group by id
func (a *AWS) GetSecurityGroup(id string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: []*string{aws.String(id)},
	}
	output, err := a.ensureEC2().DescribeSecurityGroups(input)
	if err != nil {
		return nil, exception.New(err)
	}
	if len(output.SecurityGroups) != 1 {
		return nil, exception.New(fmt.Sprintf("Incorrect number of security groups returned"))
	}
	return output.SecurityGroups[0], nil
}

// FilterSecurityGroups returns the filtered security groups
func (a *AWS) FilterSecurityGroups(filters ...*ec2.Filter) ([]*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: filters,
	}
	var output *ec2.DescribeSecurityGroupsOutput
	var err error
	groups := []*ec2.SecurityGroup{}
	for output == nil || output.NextToken != nil {
		output, err = a.ensureEC2().DescribeSecurityGroups(input)
		if err != nil {
			return nil, exception.New(err)
		}
		groups = append(groups, output.SecurityGroups...)
		input.NextToken = output.NextToken
	}
	return groups, nil
}
