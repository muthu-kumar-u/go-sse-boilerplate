package handlers

import (
	userController "github.com/nelsonin-research-org/clenz-stream/controller/user"
	"github.com/nelsonin-research-org/clenz-stream/globals"
	"github.com/nelsonin-research-org/clenz-stream/handlers"
	"github.com/nelsonin-research-org/clenz-stream/services"
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