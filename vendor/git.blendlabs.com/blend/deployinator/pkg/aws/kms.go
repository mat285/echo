package aws

import (
	"github.com/aws/aws-sdk-go/service/kms"
	exception "github.com/blend/go-sdk/exception"
)

// CreateDefaultKMSKey creates a new kms key with the default policy and returns the key metadata
func (a *AWS) CreateDefaultKMSKey() (*kms.KeyMetadata, error) {
	input := &kms.CreateKeyInput{}
	output, err := a.ensureKMS().CreateKey(input)
	if err != nil {
		return nil, exception.New(err)
	}
	return output.KeyMetadata, nil
}

// NameKMSKey creates an alias for the key. Name must start with alias/
func (a *AWS) NameKMSKey(keyID, name string) error {
	input := &kms.CreateAliasInput{
		AliasName:   &name,
		TargetKeyId: &keyID,
	}
	_, err := a.ensureKMS().CreateAlias(input)
	return exception.New(err)
}

// CreateNamedKMSKey creates a named kms key and returns it
func (a *AWS) CreateNamedKMSKey(name string) (*kms.KeyMetadata, error) {
	meta, err := a.CreateDefaultKMSKey()
	if err != nil {
		return nil, exception.New(err)
	}
	err = a.NameKMSKey(*meta.KeyId, name)
	if err != nil {
		return nil, exception.New(err)
	}
	return meta, nil
}

// GetKMSKey returns the key from the identifier. It can either be the key id, arn, alias, or alias arn
func (a *AWS) GetKMSKey(identifier string) (*kms.KeyMetadata, error) {
	input := &kms.DescribeKeyInput{
		KeyId: &identifier,
	}
	output, err := a.ensureKMS().DescribeKey(input)
	if err != nil {
		return nil, exception.New(err)
	}
	return output.KeyMetadata, nil
}
