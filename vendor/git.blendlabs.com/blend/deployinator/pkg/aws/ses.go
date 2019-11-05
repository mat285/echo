package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	exception "github.com/blend/go-sdk/exception"
)

// SendEmail uses ses to send an email
func (a *AWS) SendEmail(to, from, subject, body string) error {
	input := &ses.SendEmailInput{
		Source: &from,
		Destination: &ses.Destination{
			ToAddresses: []*string{&to},
		},
		Message: &ses.Message{
			Subject: &ses.Content{
				Charset: aws.String(defaultCharset),
				Data:    &subject,
			},
			Body: &ses.Body{
				Text: &ses.Content{
					Charset: aws.String(defaultCharset),
					Data:    &body,
				},
			},
		},
	}
	_, err := a.ensureSES().SendEmail(input)
	return exception.New(err)
}
