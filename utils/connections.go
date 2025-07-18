package utils

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/muthu-kumar-u/go-sse/globals"
	appschema "github.com/muthu-kumar-u/go-sse/models"
)

// CreateHttpClients initializes HTTP clients for the microservices with connection pool management.
func CreateHttpClients() error {
	faceAnalyzeUrl := os.Getenv("FACE_ANALYZE_SERVICE")
	userServiceUrl := os.Getenv("USER_SERVICE")

	if len(faceAnalyzeUrl) == 0 || len(userServiceUrl) == 0 {
		return errors.New("one or more service url empty")
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     60 * time.Minute,
	}

	globals.FaceAnalyzeService = &appschema.ServiceConnection{
		Client: &http.Client{
			Transport: transport,
			Timeout:   60 * time.Second,
		},
		URL: faceAnalyzeUrl,
	}

	globals.UserService = &appschema.ServiceConnection{
		Client: &http.Client{
			Transport: transport,
			Timeout:   60 * time.Second,
		},
		URL: userServiceUrl,
	}

	fmt.Println("All Services up and running")
	return nil
}
