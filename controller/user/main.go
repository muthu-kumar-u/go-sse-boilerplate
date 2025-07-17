package controller

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	constants "github.com/nelsonin-research-org/clenz-stream/const"
	"github.com/nelsonin-research-org/clenz-stream/globals"
	appschema "github.com/nelsonin-research-org/clenz-stream/models"
)

type UserController interface {
	IsUserAuthenticated(ctx context.Context, userId string) (bool, error)
}

type userControllerImpl struct {
	UserService appschema.ServiceConnection
}

func NewUserController(userService appschema.ServiceConnection) UserController {
	return &userControllerImpl{UserService: userService}
}

// user
func (s *userControllerImpl) IsUserAuthenticated(ctx context.Context, token string) (bool, error){
	reqUrl := fmt.Sprintf("%s/%s", s.UserService.URL , constants.USER_SERVICE_PATHS[0])
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return false, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := globals.FaceAnalyzeService.Client.Do(req)
	if err != nil {
		log.Printf("Error making request to user service: %v\n", err)
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("user service error %d: %s\n", resp.StatusCode, string(body))
		return false, err
	}

	return true, nil
}
