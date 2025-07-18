package services

import (
	"context"

	controller "github.com/muthu-kumar-u/go-sse/controller/user"
)

type UserService interface {
	IsUserAuthenticated(ctx context.Context, token string) (bool, error)
}

type userControllerImpl struct {
	userController controller.UserController
}

func NewUserService(c controller.UserController) UserService {
	return &userControllerImpl{userController: c}
}

func (s *userControllerImpl) IsUserAuthenticated(ctx context.Context, token string) (bool, error) {
	return s.userController.IsUserAuthenticated(ctx, token)
}
