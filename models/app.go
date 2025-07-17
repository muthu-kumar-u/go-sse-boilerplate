package appschema

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"sync"
	"time"
)

// service connection
type ServiceConnection struct {
	Client *http.Client
	URL    string
}


// request create
type ImagePayload struct {
	RawBytes        []byte
	MultipartBody   *bytes.Buffer
	MultipartWriter *multipart.Writer
	ContentType 	string
}


// rate limitation
type RateLimitEntry struct {
	Count     int
	Timestamp time.Time
}

type RequestStore struct {
	Mu       sync.Mutex
	Requests map[string]map[string]*RateLimitEntry
}