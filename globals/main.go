package globals

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/muthu-kumar-u/go-sse/events/stream"
	appschema "github.com/muthu-kumar-u/go-sse/models"
)

var AWSSession *session.Session
var UserService *appschema.ServiceConnection
var FaceAnalyzeService *appschema.ServiceConnection
var RequestStore appschema.RequestStore

// prod
var Stream *stream.StreamHub
// var Stream *sse.Server