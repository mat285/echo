package aws

import (
	"encoding/json"

	"git.blendlabs.com/blend/deployinator/pkg/core"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	exception "github.com/blend/go-sdk/exception"
)

// DoesIAMRoleExist checks whether the specified IAM role exists
func (a *AWS) DoesIAMRoleExist(name string) (bool, error) {
	role, err := a.GetIAMRole(name)
	if err != nil {
		notFound := IgnoreErrorCodes(err, iam.ErrCodeNoSuchEntityException)
		return false, notFound
	}
	return role != nil && role.RoleName != nil && *role.RoleName == name, nil
}

// GetIAMRole returns the IAM role with the given name
func (a *AWS) GetIAMRole(name string) (*iam.Role, error) {
	input := &iam.GetRoleInput{
		RoleName: &name,
	}
	output, err := a.ensureIAM().GetRole(input)
	if err != nil {
		return nil, err
	}
	return output.Role, nil
}

// CreateIAMRole creates the specified role
func (a *AWS) CreateIAMRole(name, description, policyDocument string) (string, error) {
	input := &iam.CreateRoleInput{
		RoleName:                 &name,
		Description:              &description,
		AssumeRolePolicyDocument: &policyDocument,
	}
	return a.CreateIAMRoleWithInput(input)
}

// CreateIAMRoleWithInput creates a new iam role
func (a *AWS) CreateIAMRoleWithInput(input *iam.CreateRoleInput) (string, error) {
	output, err := a.ensureIAM().CreateRole(input)
	if err != nil {
		return "", exception.New(err)
	}
	return aws.StringValue(output.Role.Arn), nil
}

// CreateInstanceProfile creates the instance profile with the given name
func (a *AWS) CreateInstanceProfile(name string) error {
	input := &iam.CreateInstanceProfileInput{
		InstanceProfileName: &name,
	}
	_, err := a.ensureIAM().CreateInstanceProfile(input)
	return exception.New(err)
}

// AddRoleToInstanceProfile adds the specified role to the profile
func (a *AWS) AddRoleToInstanceProfile(role, profile string) error {
	input := &iam.AddRoleToInstanceProfileInput{
		RoleName:            &role,
		InstanceProfileName: &profile,
	}
	_, err := a.ensureIAM().AddRoleToInstanceProfile(input)
	return exception.New(err)
}

// UpdateAssumeRolePolicy updates the assume role trust policy on the role
func (a *AWS) UpdateAssumeRolePolicy(roleName string, policy string) error {
	input := &iam.UpdateAssumeRolePolicyInput{
		RoleName:       &roleName,
		PolicyDocument: &policy,
	}
	_, err := a.ensureIAM().UpdateAssumeRolePolicy(input)
	return exception.New(err)
}

// ListRolePolicies lists the roles inline policies 
func (a *AWS) ListRolePolicies(roleName string) ([]string, error) {
	input := &iam.ListRolePoliciesInput{
		RoleName: &roleName,
	}
	policies := []string{}
	err := a.ensureIAM().ListRolePoliciesPages(input, func(out *iam.ListRolePoliciesOutput, _ bool) bool {
		policies = append(policies, core.PtrSliceToStringSlice(out.PolicyNames)...)
		return true
	})
	return policies, exception.New(err)
}

// PutRolePolicy puts the policy onto the role
func (a *AWS) PutRolePolicy(policyName, roleName, policyDocument string) error {
	input := &iam.PutRolePolicyInput{
		PolicyName:     &policyName,
		PolicyDocument: &policyDocument,
		RoleName:       &roleName,
	}
	_, err := a.ensureIAM().PutRolePolicy(input)
	return exception.New(err)
}

// DeleteRolePolicy deletes the specified policy from the role
func (a *AWS) DeleteRolePolicy(policy, role string) error {
	input := &iam.DeleteRolePolicyInput{
		PolicyName: &policy,
		RoleName:   &role,
	}
	_, err := a.ensureIAM().DeleteRolePolicy(input)
	return IgnoreErrorCodes(err, iam.ErrCodeNoSuchEntityException)
}

// DeleteRole deletes the specified role
func (a *AWS) DeleteRole(name string) error {
	input := &iam.DeleteRoleInput{
		RoleName: &name,
	}
	_, err := a.ensureIAM().DeleteRole(input)
	return exception.New(err)
}

// DeleteInstanceProfile deletes the specified instance profile
func (a *AWS) DeleteInstanceProfile(name string) error {
	input := &iam.DeleteInstanceProfileInput{
		InstanceProfileName: &name,
	}
	_, err := a.ensureIAM().DeleteInstanceProfile(input)
	return exception.New(err)
}

// DeleteRolePolicyAndProfile deletes the role and policy on the role, and the profile, ignores not found exceptions
func (a *AWS) DeleteRolePolicyAndProfile(role, profile, policy string) error {
	err := a.DeleteRolePolicy(policy, role)
	if err != nil {
		return exception.New(err)
	}
	err = IgnoreErrorCodes(a.RemoveRoleFromProfile(role, profile), iam.ErrCodeNoSuchEntityException)
	if err != nil {
		return exception.New(err)
	}
	err = IgnoreErrorCodes(a.DeleteRole(role), iam.ErrCodeNoSuchEntityException)
	if err != nil {
		return exception.New(err)
	}
	return IgnoreErrorCodes(a.DeleteInstanceProfile(profile), iam.ErrCodeNoSuchEntityException)
}

// RemoveRoleFromProfile removes the role from the instance profile
func (a *AWS) RemoveRoleFromProfile(role, profile string) error {
	input := &iam.RemoveRoleFromInstanceProfileInput{
		RoleName:            &role,
		InstanceProfileName: &profile,
	}
	_, err := a.ensureIAM().RemoveRoleFromInstanceProfile(input)
	return exception.New(err)
}

// CreateUser creates a user with the given name
func (a *AWS) CreateUser(name string) error {
	input := &iam.CreateUserInput{
		UserName: aws.String(name),
	}
	_, err := a.ensureIAM().CreateUser(input)
	return exception.New(err)
}

// DeleteUser deletes the specified user
func (a *AWS) DeleteUser(name string) error {
	err := a.DeleteAccessKeys(name)
	if err != nil {
		return exception.New(err)
	}
	err = a.DeleteUserPolicies(name)
	if err != nil {
		return exception.New(err)
	}
	input := &iam.DeleteUserInput{
		UserName: aws.String(name),
	}
	_, err = a.ensureIAM().DeleteUser(input)
	return exception.New(err)
}

// GetUser gets the user
func (a *AWS) GetUser(name string) (*iam.User, error) {
	input := &iam.GetUserInput{
		UserName: aws.String(name),
	}
	output, err := a.ensureIAM().GetUser(input)
	if err != nil {
		return nil, exception.New(err)
	}
	return output.User, nil
}

// UserExists returns if the user exists
func (a *AWS) UserExists(name string) (bool, error) {
	user, err := a.GetUser(name)
	if err != nil {
		notFound := IgnoreErrorCodes(err, iam.ErrCodeNoSuchEntityException)
		return false, notFound
	}
	return user != nil && user.UserName != nil && *user.UserName == name, nil
}

// ListAccessKeys lists the access keys for the user
func (a *AWS) ListAccessKeys(user string) ([]*iam.AccessKeyMetadata, error) {
	input := &iam.ListAccessKeysInput{
		UserName: aws.String(user),
	}
	keys := []*iam.AccessKeyMetadata{}
	err := a.ensureIAM().ListAccessKeysPages(input, func(output *iam.ListAccessKeysOutput, lastPage bool) bool {
		if output != nil && output.AccessKeyMetadata != nil {
			for _, meta := range output.AccessKeyMetadata {
				if meta != nil && meta.AccessKeyId != nil {
					keys = append(keys, meta)
				}
			}
		}
		return true
	})
	if err != nil {
		return nil, exception.New(err)
	}
	return keys, nil
}

// DeleteAccessKeys deletes all the access keys for the users
func (a *AWS) DeleteAccessKeys(user string) error {
	keys, err := a.ListAccessKeys(user)
	if err != nil {
		return exception.New(err)
	}
	for _, key := range keys {
		err = a.DeleteAccessKey(user, aws.StringValue(key.AccessKeyId))
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteAccessKey deletes the specified accesskey from the user
func (a *AWS) DeleteAccessKey(user, key string) error {
	input := &iam.DeleteAccessKeyInput{
		AccessKeyId: aws.String(key),
		UserName:    aws.String(user),
	}
	_, err := a.ensureIAM().DeleteAccessKey(input)
	return exception.New(err)
}

// CreateAccessKey generates an access key for the user
func (a *AWS) CreateAccessKey(user string) (*iam.AccessKey, error) {
	input := &iam.CreateAccessKeyInput{
		UserName: aws.String(user),
	}
	output, err := a.ensureIAM().CreateAccessKey(input)
	if err != nil {
		return nil, exception.New(err)
	}
	return output.AccessKey, nil
}

// PutUserPolicy puts the policy on the user
func (a *AWS) PutUserPolicy(user, policyName, policyDocument string) error {
	input := &iam.PutUserPolicyInput{
		UserName:       aws.String(user),
		PolicyName:     aws.String(policyName),
		PolicyDocument: aws.String(policyDocument),
	}
	_, err := a.ensureIAM().PutUserPolicy(input)
	return exception.New(err)
}

// ListUserPolicies lists the user policies
func (a *AWS) ListUserPolicies(user string) ([]string, error) {
	input := &iam.ListUserPoliciesInput{
		UserName: aws.String(user),
	}
	names := []string{}
	err := a.ensureIAM().ListUserPoliciesPages(input, func(output *iam.ListUserPoliciesOutput, last bool) bool {
		if output != nil {
			for _, name := range output.PolicyNames {
				if name != nil {
					names = append(names, *name)
				}
			}
		}
		return true
	})
	if err != nil {
		return nil, exception.New(err)
	}
	return names, nil
}

// DeleteUserPolicy deletes the policy from the user
func (a *AWS) DeleteUserPolicy(user, policy string) error {
	input := &iam.DeleteUserPolicyInput{
		PolicyName: aws.String(policy),
		UserName:   aws.String(user),
	}
	_, err := a.ensureIAM().DeleteUserPolicy(input)
	return exception.New(err)
}

// DeleteUserPolicies deletes all the policies from the user
func (a *AWS) DeleteUserPolicies(user string) error {
	policies, err := a.ListUserPolicies(user)
	if err != nil {
		return exception.New(err)
	}
	for _, policy := range policies {
		err = a.DeleteUserPolicy(user, policy)
		if err != nil {
			return exception.New(err)
		}
	}
	return nil
}

// AttachRolePolicy attaches the policy to the role
func (a *AWS) AttachRolePolicy(role, policyArn string) error {
	input := &iam.AttachRolePolicyInput{
		RoleName:  &role,
		PolicyArn: &policyArn,
	}
	_, err := a.ensureIAM().AttachRolePolicy(input)
	return exception.New(err)
}

// ListPolicies lists all the policies in the account
func (a *AWS) ListPolicies() ([]*iam.Policy, error) {
	policies := []*iam.Policy{}
	input := &iam.ListPoliciesInput{}
	err := a.ensureIAM().ListPoliciesPages(input, func(output *iam.ListPoliciesOutput, last bool) bool {
		policies = append(policies, output.Policies...)
		return true
	})
	if err != nil {
		return nil, exception.New(err)
	}
	return policies, nil
}

// GetPolicyByName returns the policy by the name if it exists
func (a *AWS) GetPolicyByName(name string) (*iam.Policy, error) {
	policies, err := a.ListPolicies()
	if err != nil {
		return nil, err
	}
	for _, policy := range policies {
		if policy != nil && policy.PolicyName != nil && *policy.PolicyName == name {
			return policy, nil
		}
	}
	return nil, NewError(iam.ErrCodeNoSuchEntityException, "Policy not found")
}

// CreatePolicy creates the policy in aws
func (a *AWS) CreatePolicy(name, policy string) error {
	input := &iam.CreatePolicyInput{
		PolicyName:     &name,
		PolicyDocument: &policy,
	}
	_, err := a.ensureIAM().CreatePolicy(input)
	return exception.New(err)
}

// DeletePolicy deletes the policy by arn
func (a *AWS) DeletePolicy(arn string) error {
	input := &iam.DeletePolicyInput{
		PolicyArn: &arn,
	}
	_, err := a.ensureIAM().DeletePolicy(input)
	return exception.New(err)
}

// JSON converts the policy document to json
func (i *IAMPolicyDocument) JSON() ([]byte, error) {
	return json.Marshal(i)
}

// JSON converts the policy document to json
func (i *IAMTrustPolicyDocument) JSON() ([]byte, error) {
	return json.Marshal(i)
}

// NewIAMAssumeRolePolicyDocument instantiates an iam assume role policy document for a service
func NewIAMAssumeRolePolicyDocument(service string) *IAMAssumeRolePolicyDocument {
	return &IAMAssumeRolePolicyDocument{
		Version: IAMPolicyVersion,
		Statement: IAMAssumeRoleStatement{
			Effect: "Allow",
			Principal: IAMPrincipal{
				Service: service,
			},
			Action: "sts:AssumeRole",
		},
	}
}
