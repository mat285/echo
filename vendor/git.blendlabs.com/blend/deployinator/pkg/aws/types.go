package aws

// AWS does not provide these types for us, so we provide them here

// IAMEffect is an effect property
type IAMEffect string

const (
	// EffectAllow sets allow on the policy
	EffectAllow IAMEffect = "Allow"
	// EffectDeny sets deny on the policy
	EffectDeny IAMEffect = "Deny"
)

// IAMPolicyDocument is the top level iam policy, composed of many statements
type IAMPolicyDocument struct {
	Version   string         `json:"Version"`
	Statement []IAMStatement `json:"Statement"`
}

// IAMTrustPolicyDocument is the top level iam policy for trust policies
type IAMTrustPolicyDocument struct {
	Version   string              `json:"Version"`
	Statement []IAMTrustStatement `json:"Statement"`
}

// IAMStatement is an iam set of permissions
type IAMStatement struct {
	Sid       string       `json:"Sid" yaml:"Sid"`
	Effect    IAMEffect    `json:"Effect" yaml:"Effect"`
	Action    []string     `json:"Action" yaml:"Action"`
	Resource  []string     `json:"Resource" yaml:"Resource"`
	Condition IAMCondition `json:"Condition,omitempty" yaml:"Condition,omitempty"`
}

// IAMTrustStatement is an iam set of permissions (for trust documents)
type IAMTrustStatement struct {
	Sid       string            `json:"Sid" yaml:"Sid"`
	Effect    IAMEffect         `json:"Effect" yaml:"Effect"`
	Principal map[string]string `json:"Principal,omitempty" yaml:"Principal,omitempty"`
	Action    string            `json:"Action" yaml:"Action"`
}

// IAMCondition is a condition object
type IAMCondition map[IAMConditionOperator]map[string]string

// IAMConditionOperator is an operator for an iam condition
type IAMConditionOperator string

// IAMAssumeRolePolicyDocument is the top level iam policy for assume role policy types
type IAMAssumeRolePolicyDocument struct {
	Version   string                 `json:"Version" yaml:"Version"`
	Statement IAMAssumeRoleStatement `json:"Statement" yaml:"Statement"`
}

// S3BucketPolicyDocument is a policy document for an S3 bucket
type S3BucketPolicyDocument struct {
	Version   string                    `json:"Version" yaml:"Version"`
	Statement []S3BucketPolicyStatement `json:"Statement" yaml:"Statement"`
}

// S3BucketPolicyStatement is an S3 rule for a bucket policy
type S3BucketPolicyStatement struct {
	Effect    string      `json:"Effect" yaml:"Effect"`
	Principal S3Principal `json:"Principal" yaml:"Principal"`
	Action    string      `json:"Action" yaml:"Action"`
	Resource  string      `json:"Resource" yaml:"Resource"`
}

// IAMAssumeRoleStatement is an iam rule to assume roles
type IAMAssumeRoleStatement struct {
	Effect    string       `json:"Effect" yaml:"Effect"`
	Principal IAMPrincipal `json:"Principal" yaml:"Principal"`
	Action    string       `json:"Action" yaml:"Action"`
}

// IAMPrincipal is the principal for this statement
type IAMPrincipal struct {
	Service string `json:"Service" yaml:"Service"`
}

// S3Principal is the principal for an s3 policy statement
type S3Principal struct {
	AWS string `json:"AWS" yaml:"AWS"`
}

// ECRAuth is the auth for ecr
type ECRAuth struct {
	Registry string
	Username string
	Password string
}
