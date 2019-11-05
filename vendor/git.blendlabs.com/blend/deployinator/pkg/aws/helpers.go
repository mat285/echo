package aws

import "fmt"

// AssumeRolePolicy is a policy to attach to a role allowing it to be assumed
func AssumeRolePolicy(roleArn string) *IAMPolicyDocument {
	return &IAMPolicyDocument{
		Version: IAMPolicyVersion,
		Statement: []IAMStatement{
			AssumeRolePolicyStatement(roleArn),
		},
	}

}

// AssumeRolePolicyStatement is a statement allowing assume role to the arn
func AssumeRolePolicyStatement(roleArn string) IAMStatement {
	return IAMStatement{
		Sid:      "AssumeRole",
		Effect:   EffectAllow,
		Resource: []string{roleArn},
		Action:   []string{"sts:AssumeRole"},
	}
}

// AssumeRoleTrustStatement is the trust statement to attach to a role to allow it to be assume by another role
func AssumeRoleTrustStatement(sid, roleArn string) IAMTrustStatement {
	return IAMTrustStatement{
		Sid:    sid,
		Effect: EffectAllow,
		Principal: map[string]string{
			PrincipalAWS: roleArn,
		},
		Action: "sts:AssumeRole",
	}
}

// AssumeRoleTrustEC2Statement is the statement allowing ec2 to assume the role
func AssumeRoleTrustEC2Statement() IAMTrustStatement {
	return IAMTrustStatement{
		Effect: EffectAllow,
		Principal: map[string]string{
			PrincipalService: ServiceEC2,
		},
		Action: "sts:AssumeRole",
	}
}

// ArnPrefix is the prefix to all aws arns
func ArnPrefix() string {
	return "arn:aws:"
}

// IAMArnPrefix is the prefix to all iam arns
func IAMArnPrefix(accountID string) string {
	return fmt.Sprintf("%siam::%s:", ArnPrefix(), accountID)
}

// IAMRoleArn is the arn for a role with the given name
func IAMRoleArn(accountID, roleName string) string {
	return fmt.Sprintf("%srole/%s", IAMArnPrefix(accountID), roleName)
}

// UpsertIAMStatementByID updates the policy by the statement id, or appends it to the policy if it doesn't exist
func UpsertIAMStatementByID(statement IAMStatement, policy IAMPolicyDocument) IAMPolicyDocument {
	idx := -1
	for i, s := range policy.Statement {
		if s.Sid == statement.Sid {
			idx = i
			break
		}
	}
	if idx < 0 || idx >= len(policy.Statement) {
		policy.Statement = append(policy.Statement, statement)
	} else {
		policy.Statement[idx] = statement
	}
	return policy
}

// UpsertIAMTrustStatementByID updates the policy by the statement id, or appends it to the policy if it doesn't exist
func UpsertIAMTrustStatementByID(statement IAMTrustStatement, policy IAMTrustPolicyDocument) IAMTrustPolicyDocument {
	idx := -1
	for i, s := range policy.Statement {
		if s.Sid == statement.Sid {
			idx = i
			break
		}
	}
	if idx < 0 || idx >= len(policy.Statement) {
		policy.Statement = append(policy.Statement, statement)
	} else {
		policy.Statement[idx] = statement
	}
	return policy
}
