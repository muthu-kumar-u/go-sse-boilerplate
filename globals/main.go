package globals

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/nelsonin-research-org/clenz-stream/events/stream"
	appschema "github.com/nelsonin-research-org/clenz-stream/models"
)

var AWSSession *session.Session
var UserService *appschema.ServiceConnection
var FaceAnalyzeService *appschema.ServiceConnection
var RequestStore appschema.RequestStore

// prod
var Stream *stream.StreamHub
// var Stream *sse.Server