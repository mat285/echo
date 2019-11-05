package aws

import (
	"encoding/json"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	exception "github.com/blend/go-sdk/exception"
)

// DoesBucketExist checks whether the specified bucket exists
func (a *AWS) DoesBucketExist(name string) (bool, error) {
	one := int64(1) //just for kicks
	input := &s3.ListObjectsInput{
		Bucket:  &name,
		MaxKeys: &one,
	}
	_, err := a.ensureS3().ListObjects(input)
	if err != nil {
		err2 := IgnoreErrorCodes(err, s3.ErrCodeNoSuchBucket)
		return false, exception.New(err2)
	}
	return true, nil
}

// GetBucketPolicy retrieves the policy for the given bucket
func (a *AWS) GetBucketPolicy(bucketName string) (*string, error) {
	input := &s3.GetBucketPolicyInput{
		Bucket: &bucketName,
	}
	getPolicyOutput, err := a.ensureS3().GetBucketPolicy(input)
	return getPolicyOutput.Policy, err
}

// CreateBucket creates a bucket with the given name
func (a *AWS) CreateBucket(name string) error {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(name),
	}
	_, err := a.ensureS3().CreateBucket(input)
	if err != nil {
		return exception.New(err)
	}
	err = a.ensureS3().WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(name),
	})
	if err != nil {
		return exception.New(err)
	}
	_, err = a.ensureS3().PutBucketEncryption(&s3.PutBucketEncryptionInput{
		Bucket: aws.String(name),
		ServerSideEncryptionConfiguration: &s3.ServerSideEncryptionConfiguration{
			Rules: []*s3.ServerSideEncryptionRule{
				{
					ApplyServerSideEncryptionByDefault: &s3.ServerSideEncryptionByDefault{
						SSEAlgorithm: aws.String(s3.ServerSideEncryptionAes256),
					},
				},
			},
		},
	})
	return exception.New(err)
}

// PutBucketPolicy puts a bucket policy onto the given bucket
func (a *AWS) PutBucketPolicy(bucketName string, policy string) error {
	policyInput := s3.PutBucketPolicyInput{
		Bucket: &bucketName,
		Policy: &policy,
	}
	_, err := a.ensureS3().PutBucketPolicy(&policyInput)
	return exception.New(err)
}

// JSON converts the policy document to json
func (i *S3BucketPolicyDocument) JSON() ([]byte, error) {
	return json.Marshal(i)
}

// DeleteBucket deletes a bucket with the given name
func (a *AWS) DeleteBucket(name string) error {
	client := a.ensureS3()
	bucket := aws.String(name)
	// We must delete all objects inside before we can delete a bucket.
	// http://docs.aws.amazon.com/AmazonS3/latest/dev/delete-or-empty-bucket.html#delete-bucket-awssdks
	var deleteObjectsErr error
	err := client.ListObjectsPages(&s3.ListObjectsInput{Bucket: bucket}, func(output *s3.ListObjectsOutput, lastPage bool) bool {
		var objectIds []*s3.ObjectIdentifier
		for _, object := range output.Contents {
			objectIds = append(objectIds, &s3.ObjectIdentifier{
				Key: object.Key,
			})
		}
		_, err := client.DeleteObjects(&s3.DeleteObjectsInput{
			Bucket: bucket,
			Delete: &s3.Delete{Objects: objectIds},
		})
		if err != nil {
			deleteObjectsErr = err
			return false
		}
		return true
	})

	if deleteObjectsErr != nil {
		return deleteObjectsErr
	}

	input := &s3.DeleteBucketInput{
		Bucket: bucket,
	}
	_, err = a.ensureS3().DeleteBucket(input)
	return exception.New(err)
}

// InitializeBucket puts an object on the bucket to work around a registry bug
// https://github.com/docker/distribution/issues/2292
func (a *AWS) InitializeBucket(name string) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(name),
		Key:    aws.String("DELETEME"),
	}
	_, err := a.ensureS3().PutObject(input)
	return exception.New(err)
}

// ListObjects lists the objects in the bucket
func (a *AWS) ListObjects(bucket, prefix string) ([]string, error) {
	input := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}
	objs := []string{}
	err := a.ensureS3().ListObjectsPages(input, func(output *s3.ListObjectsOutput, last bool) bool {
		for _, o := range output.Contents {
			if o.Key != nil {
				objs = append(objs, *o.Key)
			}
		}
		return true
	})
	return objs, exception.New(err)
}

// GetObject gets the object with the key from the bucket
func (a *AWS) GetObject(bucket, key string) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	output, err := a.ensureS3().GetObject(input)
	if err != nil {
		return nil, exception.New(err)
	}
	return output.Body, nil
}

// PutObject puts the data into the bucket with the given key
func (a *AWS) PutObject(bucket, key string, data io.ReadSeeker) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   data,
	}
	_, err := a.ensureS3().PutObject(input)
	return exception.New(err)
}

// DeleteObject deletes the object from s3
func (a *AWS) DeleteObject(bucket, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	_, err := a.ensureS3().DeleteObject(input)
	return exception.New(err)
}

// GetBucketLifecycleRules returns the lifecycle rules for the bucket
func (a *AWS) GetBucketLifecycleRules(bucket string) ([]*s3.LifecycleRule, error) {
	input := &s3.GetBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucket),
	}
	output, err := a.ensureS3().GetBucketLifecycleConfiguration(input)
	if err != nil {
		return nil, exception.New(err)
	}
	return output.Rules, nil
}

// GetBucketLifecycleRulesIfExist gets the bucket rules if they exist, and returns nil if there are none
func (a *AWS) GetBucketLifecycleRulesIfExist(bucket string) ([]*s3.LifecycleRule, error) {
	rules, err := a.GetBucketLifecycleRules(bucket)
	if err != nil {
		return nil, IgnoreErrorCodes(err, ErrCodeNoSuchLifecycleConfiguration)
	}
	return rules, nil
}

// PutBucketLifecycleRules replaces the buckets current rules with the specified ones
func (a *AWS) PutBucketLifecycleRules(bucket string, rules []*s3.LifecycleRule) error {
	input := &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucket),
		LifecycleConfiguration: &s3.BucketLifecycleConfiguration{
			Rules: rules,
		},
	}
	_, err := a.ensureS3().PutBucketLifecycleConfiguration(input)
	return exception.New(err)
}

// GetBucketReplicationRules returns the replication rules for the bucket
func (a *AWS) GetBucketReplicationRules(bucket string) ([]*s3.ReplicationRule, error) {
	input := &s3.GetBucketReplicationInput{
		Bucket: aws.String(bucket),
	}
	output, err := a.ensureS3().GetBucketReplication(input)
	if err != nil {
		return nil, exception.New(err)
	}
	if output.ReplicationConfiguration == nil {
		return nil, exception.New(ErrCodeNoSuchReplicationConfiguration)
	}
	return output.ReplicationConfiguration.Rules, nil
}

// GetBucketReplicationRulesIfExist gets the bucket replication rules if exist, and returns nil if there are none
func (a *AWS) GetBucketReplicationRulesIfExist(bucket string) ([]*s3.ReplicationRule, error) {
	rules, err := a.GetBucketReplicationRules(bucket)
	if err != nil {
		// the doc says the former is the error code returned, the latter is what we found in practice. so we just ignore both.
		return nil, IgnoreErrorCodes(err, ErrCodeNoSuchReplicationConfiguration, ErrCodeReplicationConfigurationNotFoundError)
	}
	return rules, nil
}

// PutBucketReplicationRules put bucket replication rules to the configuration
func (a *AWS) PutBucketReplicationRules(bucket, roleARN string, rules []*s3.ReplicationRule) error {
	input := &s3.PutBucketReplicationInput{
		Bucket: aws.String(bucket),
		ReplicationConfiguration: &s3.ReplicationConfiguration{
			Role:  aws.String(roleARN),
			Rules: rules,
		},
	}
	_, err := a.ensureS3().PutBucketReplication(input)
	return exception.New(err)
}

// PutBucketVersioning replaces the buckets current rules with the specified ones
func (a *AWS) PutBucketVersioning(bucket string) error {
	input := &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucket),
		VersioningConfiguration: &s3.VersioningConfiguration{
			MFADelete: aws.String(s3.MFADeleteStatusDisabled),
			Status:    aws.String(s3.BucketVersioningStatusEnabled),
		},
	}
	_, err := a.ensureS3().PutBucketVersioning(input)
	return exception.New(err)
}
