package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nelsonin-research-org/clenz-stream/events/stream"
	"github.com/nelsonin-research-org/clenz-stream/globals"
	"github.com/nelsonin-research-org/clenz-stream/handlers"
	app "github.com/nelsonin-research-org/clenz-stream/handlers/data"
	"github.com/nelsonin-research-org/clenz-stream/middleware"
	"github.com/nelsonin-research-org/clenz-stream/utils"
)

var ginApp *gin.Engine
var ginLambda *ginadapter.GinLambdaV2
var streamHandler  *handlers.StreamHandler

// func mainHandler(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLStreamingResponse, error) {
// 	var res events.LambdaFunctionURLStreamingResponse
// 	var err error

// 	switch {
// 	case req.RequestContext.HTTP.Method == http.MethodGet && strings.HasPrefix(req.RawPath, "/api/v1/facelog"):
// 		res, err = streamHandler.FaceLogStreamLambda(ctx, req)

// 	case req.RequestContext.HTTP.Method == http.MethodPost && strings.HasPrefix(req.RawPath, "/api/v1/facelog/upload"):
// 		res, err = streamHandler.LogUserFaceLambda(ctx, req)

// 	default:
// 		res, err = stream.SSEErrorStream(http.StatusNotFound, "not found")
// 	}

// 	ct := res.Headers["Content-Type"]
// 	if ct == "" {
// 		ct = "(none set)"
// 	}
// 	fmt.Printf("🔄 Final Response Content-Type: %s\n", ct)

// 	return res, err
// }

func lambdaHandler(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLStreamingResponse, error) {
    if ginLambda == nil {
        ginLambda = ginadapter.NewV2(ginApp)
    }

    // Convert Lambda URL request to APIGateway request
    apiReq := events.APIGatewayV2HTTPRequest{
        Version:               "2.0",
        RouteKey:              "",
        RawPath:               req.RawPath,
        RawQueryString:        req.RawQueryString,
        Cookies:               req.Cookies,
        Headers:               req.Headers,
        QueryStringParameters: req.QueryStringParameters,
        RequestContext: events.APIGatewayV2HTTPRequestContext{
            HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
                Method: req.RequestContext.HTTP.Method,
                Path:   req.RawPath,
            },
            DomainName: req.RequestContext.DomainName,
        },
        Body:            req.Body,
        IsBase64Encoded: req.IsBase64Encoded,
    }

    resp, err := ginLambda.ProxyWithContext(ctx, apiReq)
    if err != nil {
        return events.LambdaFunctionURLStreamingResponse{}, err
    }


	 // 🔍 Log the response content
    fmt.Println("=== Lambda Response ===")
    fmt.Printf("Status Code: %d\n", resp.StatusCode)
    fmt.Printf("Content-Type: %s\n", resp.Headers["Content-Type"])
    fmt.Printf("Body:\n%s\n", resp.Body)

    return events.LambdaFunctionURLStreamingResponse{
        StatusCode:        resp.StatusCode,
        Headers:           resp.Headers,
        Body:              strings.NewReader(resp.Body),
        Cookies:           resp.Cookies,
    }, nil
}

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

func main() {
	if err := Init(); err != nil {
		log.Fatalf("Initialization error: %v", err)
	}
	defer os.Exit(1)

	handlers := app.LoadAppHandlers()
	streamHandler = handlers.StreamHandler

	// Initialize Gin
	ginApp = gin.New()
	
	// Middleware
	ginApp.Use(gin.Logger(), gin.Recovery(), utils.GetCorsConfig())

	version := os.Getenv("APP_VERSION")
	port := os.Getenv("APP_PORT")
	production := os.Getenv("PRODUCTION") == "true"

	api := ginApp.Group("/api/" + version)
	{
		api.GET("/facelog", middleware.AuthMiddleware(handlers.StreamHandler.UserService), handlers.StreamHandler.FaceLogStream)
		api.POST("/facelog/upload", middleware.AuthMiddleware(handlers.StreamHandler.UserService), handlers.StreamHandler.LogUserFace)
	}

	// 404 handler
	ginApp.NoRoute(middleware.PathNotFound())

	globals.Stream = stream.NewStreamHub()

	if production {
		log.Println("Running as Lambda function...")
		lambda.Start(lambdaHandler)
	} else {
		log.Printf("Starting local server on :%s\n", port)
		log.Fatal(ginApp.Run(":" + port))
	}
}