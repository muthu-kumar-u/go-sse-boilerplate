package utils

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/muthu-kumar-u/go-sse/globals"
)

func AWSSessionConfigure() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String(os.Getenv("AWS_SERVICE_REGION")),
		},
		SharedConfigState: session.SharedConfigEnable,
	}))
	globals.AWSSession = sess
}