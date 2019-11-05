package aws

import (
	"sync"
	"time"

	"git.blendlabs.com/blend/deployinator/pkg/core"
	"git.blendlabs.com/blend/deployinator/pkg/logging"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	exception "github.com/blend/go-sdk/exception"
)

const (
	defaultAWSRegion = "us-1-east"
)

var (
	defaultAWS  *AWS
	defaultLock sync.Mutex
)

// AWS encapsulates the aws services
type AWS struct {
	session *session.Session
	config  *aws.Config
	ec2     ec2iface.EC2API
	ecr     ecriface.ECRAPI
	elb     elbiface.ELBAPI
	elbv2   elbv2iface.ELBV2API
	route53 route53iface.Route53API
	asg     autoscalingiface.AutoScalingAPI
	iam     iamiface.IAMAPI
	s3      s3iface.S3API
	cf      cloudformationiface.CloudFormationAPI
	kms     kmsiface.KMSAPI
	ses     sesiface.SESAPI

	log Logger
}

// Logger is a logger to call for DoAction
type Logger interface {
	Infof(string, ...interface{})
	Error(error) error
}

type emptyLog struct {
}

func (e *emptyLog) Infof(fmt string, args ...interface{}) {
}
func (e *emptyLog) Error(err error) error {
	return nil
}

// Default gets the default aws
func Default() *AWS {
	return defaultAWS
}

// SetDefault sets the default aws
func SetDefault(aws *AWS) {
	defaultLock.Lock()
	defer defaultLock.Unlock()
	defaultAWS = aws
}

// New creates a new aws service for this config
func New(config *aws.Config) *AWS {
	a := &AWS{
		config: config,
		log:    &emptyLog{},
	}
	a.ensureSession()
	return a
}

// NewForRegion creates a new aws service for this region
func NewForRegion(region string) *AWS {
	return New(&aws.Config{Region: &region})
}

// SetLogger sets the logger for this aws instance
func (a *AWS) SetLogger(log Logger) {
	if log == nil {
		log = &emptyLog{}
	}
	a.log = log
}

func (a *AWS) ensureConfig() *aws.Config {
	if a.config == nil {
		a.config = &aws.Config{Region: aws.String(defaultAWSRegion)}
	}
	return a.config
}

func (a *AWS) ensureSession() *session.Session {
	if a.session == nil {
		a.session = session.Must(session.NewSessionWithOptions(session.Options{
			AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
			SharedConfigState:       session.SharedConfigEnable,
		}))
		a.session.Handlers.Send.PushBack(a.getLoggingHandler())
	}
	return a.session
}

func (a *AWS) ensureEC2() ec2iface.EC2API {
	if a.ec2 == nil {
		a.ec2 = ec2.New(a.ensureSession(), a.ensureConfig())
	}
	return a.ec2
}

func (a *AWS) ensureECR() ecriface.ECRAPI {
	if a.ecr == nil {
		a.ecr = ecr.New(a.ensureSession(), a.ensureConfig())
	}
	return a.ecr
}

func (a *AWS) ensureELBV1() elbiface.ELBAPI {
	if a.elb == nil {
		a.elb = elb.New(a.ensureSession(), a.ensureConfig())
	}
	return a.elb
}

func (a *AWS) ensureELBV2() elbv2iface.ELBV2API {
	if a.elbv2 == nil {
		a.elbv2 = elbv2.New(a.ensureSession(), a.ensureConfig())
	}
	return a.elbv2
}

func (a *AWS) ensureRoute53() route53iface.Route53API {
	if a.route53 == nil {
		a.route53 = route53.New(a.ensureSession(), a.ensureConfig())
	}
	return a.route53
}

func (a *AWS) ensureASG() autoscalingiface.AutoScalingAPI {
	if a.asg == nil {
		a.asg = autoscaling.New(a.ensureSession(), a.ensureConfig())
	}
	return a.asg
}

func (a *AWS) ensureIAM() iamiface.IAMAPI {
	if a.iam == nil {
		a.iam = iam.New(a.ensureSession(), a.ensureConfig())
	}
	return a.iam
}

func (a *AWS) ensureS3() s3iface.S3API {
	if a.s3 == nil {
		a.s3 = s3.New(a.ensureSession(), a.ensureConfig())
	}
	return a.s3
}

func (a *AWS) ensureKMS() kmsiface.KMSAPI {
	if a.kms == nil {
		a.kms = kms.New(a.ensureSession(), a.ensureConfig())
	}
	return a.kms
}

func (a *AWS) ensureSES() sesiface.SESAPI {
	if a.ses == nil {
		a.ses = ses.New(a.ensureSession(), a.ensureConfig())
	}
	return a.ses
}

func (a *AWS) getLoggingHandler() func(*request.Request) {
	return func(r *request.Request) {
		if r != nil {
			logging.LogAwsRequest(r.HTTPRequest.URL.String(), r.HTTPRequest.Method)
		}
	}
}

// IgnoreErrorCodes turns the specified error code errors into nil
func IgnoreErrorCodes(err error, codes ...string) error {
	if err != nil {
		for _, code := range codes {
			if ErrorCode(err) == code {
				return nil
			}
		}
		return exception.New(err)
	}
	return nil
}

// ErrorCode returns the code of the inner aws error
func ErrorCode(err error) string {
	err = core.ExceptionUnwrap(err)
	if aerr, ok := err.(awserr.Error); ok {
		return aerr.Code()
	}
	return ErrUnknownCode
}

// NewError creates a new exception wrapped aws error with optional original error
func NewError(code, message string, origin ...error) error {
	err := error(nil)
	if len(origin) > 0 {
		err = origin[0]
	}
	return exception.New(awserr.New(code, message, err))
}

// Action is a function to run
type Action func() error

// IgnoreError turns ignored errors into nil
type IgnoreError func(error) error

// DoAction runs the action until nil is returned, a non-ignored error is returned, or a timeout occurs
func (a *AWS) DoAction(action Action, ignore IgnoreError, sleep time.Duration, timeout time.Duration) error {
	start := time.Now()
	do := true
	for do {
		err := action()
		if err == nil {
			return nil
		}
		if a.log != nil {
			a.log.Error(err)
		}
		err = ignore(err)
		if err != nil {
			return exception.New(err)
		}
		time.Sleep(sleep)
		if time.Since(start) >= timeout {
			return exception.New("Timeout performing action")
		}
	}
	return exception.New("An unknown error occurred")
}

// IgnoreValidationErrors is a function that ignores validation errors
func IgnoreValidationErrors(err error) error {
	return IgnoreErrorCodes(err, ErrCodeValidationError)
}

// IgnoreClientTokenErrors is a function that ignores aws creds error from eventual consistency
func IgnoreClientTokenErrors(err error) error {
	return IgnoreErrorCodes(err, ErrCodeInvalidClientTokenID)
}
