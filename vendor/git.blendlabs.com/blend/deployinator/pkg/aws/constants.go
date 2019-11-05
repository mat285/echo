package aws

import "time"

const (
	// TagSubnetAccessibilityTypeKey is a tag that specifies if this subnet is public
	TagSubnetAccessibilityTypeKey = "accessibility"
	// SubnetTypePrivate determines if this subnet is private
	SubnetTypePrivate = "private"
	// SubnetTypePublic determines if this subnet is public
	SubnetTypePublic = "public"

	// TagEBSVolumeNameKey is the tag specifying the name of a ebs volume
	TagEBSVolumeNameKey = "k8s/name"

	// ErrCodeAccessDenied means we were Denied Access
	ErrCodeAccessDenied = "AccessDenied"
	// ErrCodeInvalidClientTokenID is the error code when no credentials exist
	ErrCodeInvalidClientTokenID = "InvalidClientTokenId"
	// ErrCodeNoRecordFound is the error code when no record exists
	ErrCodeNoRecordFound = "NoRecordFound"
	// ErrCodeValidationError is a validation error
	ErrCodeValidationError = "ValidationError"
	// ErrCodeNotFound is a not found error
	ErrCodeNotFound = "NotFound"
	// ErrUnknownCode is a code for when an error code is unknown
	ErrUnknownCode = "UnknownErrorCode"
	// ErrCodeNoSuchLifecycleConfiguration is an error code for when there is no lifecycle config on a bucket
	ErrCodeNoSuchLifecycleConfiguration = "NoSuchLifecycleConfiguration"
	// ErrCodeNoSuchReplicationConfiguration is an error code for when there is no replication config on a bucket
	ErrCodeNoSuchReplicationConfiguration = "NoSuchReplicationConfiguration"
	// ErrCodeReplicationConfigurationNotFoundError is an error code for when there is no replication config on a bucket
	ErrCodeReplicationConfigurationNotFoundError = "ReplicationConfigurationNotFoundError"
	// ErrCodeDuplicateIPPermission is an error for duplicate permissions
	ErrCodeDuplicateIPPermission = "InvalidPermission.Duplicate"
	// ErrCodeIPPermissionNotFound is an error for permissions not found
	ErrCodeIPPermissionNotFound = "InvalidPermission.NotFound"

	// EnvVarAWSRegion is the aws region
	EnvVarAWSRegion = "AWS_REGION"
	// EnvVarAWSAccessKeyID is the access key
	EnvVarAWSAccessKeyID = "AWS_ACCESS_KEY_ID"
	// EnvVarAWSSecretAccessKey is the secret key
	EnvVarAWSSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	// EnvVarAWSSessionToken is the session token
	EnvVarAWSSessionToken = "AWS_SESSION_TOKEN"
	// EnvVarAWSDefaultRegion is the default aws region
	EnvVarAWSDefaultRegion = "AWS_DEFAULT_REGION"
	// EnvVarAWSPolicyARN is the aws policy arn
	EnvVarAWSPolicyARN = "AWS_POLICY_ARN"

	// STSHost is the host for an sts call
	STSHost = "sts.amazonaws.com"
	// STSScheme is the scheme of the sts call
	STSScheme = "https"
	// STSURL is the url of the sts call
	STSURL = STSScheme + "://" + STSHost
	// STSGetIdenityBody is the body of the post request
	STSGetIdenityBody = "Action=GetCallerIdentity&Version=2011-06-15"

	// IAMPolicyVersion is the current aws iam policy version
	IAMPolicyVersion = "2012-10-17"

	// ServiceEC2 is a service
	ServiceEC2 = "ec2.amazonaws.com"
	// ServiceS3 is a service
	ServiceS3 = "s3.amazonaws.com"

	// ELBDualStackPrefix is the dual stack prefix for an elb
	ELBDualStackPrefix = "dualstack."

	// FilterTagKey is the key of a tag
	FilterTagKey = "tag-key"
	// FilterTagValue is the value of a tag
	FilterTagValue = "tag-value"

	// ConditionStringEquals denotes string equals comparison
	ConditionStringEquals IAMConditionOperator = "StringEquals"

	// SecurityGroupAllProtocols is the constant for all protocols in the sg
	SecurityGroupAllProtocols = "-1"

	// KMSAliasPrefix is the prefix for all kms aliases
	KMSAliasPrefix = "alias/"

	// PrincipalAWS is the principal for assume role arns
	PrincipalAWS = "AWS"
	// PrincipalService is the amazon service principal for trust
	PrincipalService = "Service"
)

var (
	// IAMReadyInterval is the default ready interval for iam credentials
	IAMReadyInterval = 1 * time.Second
	// IAMReadyTimeout is the default ready timeout for iam credentials
	IAMReadyTimeout = 10 * time.Minute
)

const (
	defaultCharset = "UTF-8"
)
