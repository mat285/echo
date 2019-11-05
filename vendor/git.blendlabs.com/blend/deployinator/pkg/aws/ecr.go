package aws

import (
	"encoding/base64"
	"fmt"
	"strings"

	"git.blendlabs.com/blend/deployinator/pkg/core"
	"github.com/aws/aws-sdk-go/service/ecr"
	exception "github.com/blend/go-sdk/exception"
)

// GetECRAuthorization returns an authorization token for the ecr registrys
func (a *AWS) GetECRAuthorization(accountIDs ...string) ([]*ECRAuth, error) {
	input := &ecr.GetAuthorizationTokenInput{
		RegistryIds: core.PtrSliceFromStringSlice(accountIDs),
	}
	output, err := a.ensureECR().GetAuthorizationToken(input)
	if err != nil {
		return nil, exception.New(err)
	}
	return decodeECRAuths(output.AuthorizationData)
}

func decodeECRAuths(authData []*ecr.AuthorizationData) ([]*ECRAuth, error) {
	ret := []*ECRAuth{}
	for _, auth := range authData {
		decoded, err := decodeECRAuthData(auth)
		if err != nil {
			return nil, err
		}
		ret = append(ret, decoded)
	}
	return ret, nil
}

func decodeECRAuthData(auth *ecr.AuthorizationData) (*ECRAuth, error) {
	if auth == nil || auth.AuthorizationToken == nil || auth.ProxyEndpoint == nil {
		return nil, exception.New(fmt.Errorf("Missing token or endpoint"))
	}
	bytes, err := base64.StdEncoding.DecodeString(*auth.AuthorizationToken)
	if err != nil {
		return nil, exception.New(err)
	}
	parts := strings.SplitN(string(bytes), ":", 2)
	if len(parts) != 2 {
		return nil, exception.New(fmt.Errorf("Invalid format"))
	}
	username, password := parts[0], parts[1]
	return &ECRAuth{
		Registry: *auth.ProxyEndpoint,
		Username: username,
		Password: password,
	}, nil
}
