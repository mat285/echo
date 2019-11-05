package aws

import (
	"net/http"
	"strings"

	awsauth "github.com/smartystreets/go-aws-auth"
)

// GetCallerIdentitySignedRequest gets a signed get caller identity request
func (a *AWS) GetCallerIdentitySignedRequest() (*http.Request, error) {
	body := strings.NewReader(STSGetIdenityBody)
	req, err := http.NewRequest("POST", STSURL, body)
	if err != nil {
		return nil, err
	}
	return awsauth.Sign4(req), nil
}
