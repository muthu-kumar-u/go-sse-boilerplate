package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/muthu-kumar-u/go-sse/events/stream"
	"github.com/muthu-kumar-u/go-sse/globals"
	"github.com/muthu-kumar-u/go-sse/handlers"
	app "github.com/muthu-kumar-u/go-sse/handlers/data"
	"github.com/muthu-kumar-u/go-sse/middleware"
	"github.com/muthu-kumar-u/go-sse/utils"
)

var streamHandler *handlers.StreamHandler

func Init() error {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	utils.AWSSessionConfigure()

	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	if err := utils.CreateHttpClients(); err != nil {
		log.Printf("Error while creating HTTP client pool: %v", err)
	}

	return nil
}

func lambdaHandler(ctx context.Context, req events.LambdaFunctionURLRequest) (*events.LambdaFunctionURLStreamingResponse, error) {
	switch {
	case req.RequestContext.HTTP.Method == http.MethodPost && strings.HasPrefix(req.RawPath, "/api/v1/facelog/upload"):
		return streamHandler.LogUserFaceLambda(ctx, req)
		
	case req.RequestContext.HTTP.Method == http.MethodOptions :
		return &events.LambdaFunctionURLStreamingResponse{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization",
			},
			Body: strings.NewReader(""),
		}, nil
	
	default:
		return &events.LambdaFunctionURLStreamingResponse{
			StatusCode: http.StatusNotFound,
			Body:       strings.NewReader("Not found"),
		}, nil
	}
}

func main() {
	if err := Init(); err != nil {
		log.Fatalf("Initialization error: %v", err)
	}
	defer os.Exit(1)

	handlers := app.LoadAppHandlers()
	streamHandler = handlers.StreamHandler
	globals.Stream = stream.NewStreamHub()

	production := os.Getenv("PRODUCTION") == "true"
	if production {
		log.Println("Running as Lambda function...")
		lambda.Start(lambdaHandler)
	} else {
		port := os.Getenv("APP_PORT")
		log.Printf("Starting local server on :%s\n", port)
		
		// Local Gin setup
		ginApp := gin.New()
		ginApp.Use(gin.Logger(), gin.Recovery(), utils.GetCorsConfig())
		
		version := os.Getenv("APP_VERSION")
		api := ginApp.Group("/api/" + version)
		{
			api.GET("/facelog", middleware.AuthMiddleware(handlers.StreamHandler.UserService), handlers.StreamHandler.FaceLogStream)
			api.POST("/facelog/upload", middleware.AuthMiddleware(handlers.StreamHandler.UserService), handlers.StreamHandler.LogUserFace)
		}
		ginApp.NoRoute(middleware.PathNotFound())
		
		log.Fatal(ginApp.Run(":" + port))
	}
}