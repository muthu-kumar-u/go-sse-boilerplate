package handlers

import (
	userController "github.com/muthu-kumar-u/go-sse/controller/user"
	"github.com/muthu-kumar-u/go-sse/globals"
	"github.com/muthu-kumar-u/go-sse/handlers"
	"github.com/muthu-kumar-u/go-sse/services"
)

type AppHandlers struct {
	StreamHandler   *handlers.StreamHandler
}

func LoadAppHandlers() *AppHandlers {
	// user
	userController := userController.NewUserController(*globals.UserService)
	userService := services.NewUserService(userController)


	return &AppHandlers{
		StreamHandler:   handlers.NewFaceAnalyzeHandler(userService),
	}
}