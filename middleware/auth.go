package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/muthu-kumar-u/go-sse/globals"
	"github.com/muthu-kumar-u/go-sse/message"
	"github.com/muthu-kumar-u/go-sse/services"
)

func AuthMiddleware(userService services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, message.ReturnMessage(http.StatusUnauthorized))
			c.Abort()
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusBadRequest, message.ReturnMessage(http.StatusBadRequest))
			c.Abort()
			return
		}

		tokenString := parts[1]

		if globals.UserService == nil {
			fmt.Println("UserService not initialized")
			c.JSON(http.StatusInternalServerError, message.ReturnMessage(http.StatusInternalServerError))
			c.Abort()
			return
		}

		verified, err := userService.IsUserAuthenticated(context.TODO(), tokenString)
		if err != nil {
			fmt.Println("error while authenticate user", err.Error())
			c.JSON(http.StatusInternalServerError, message.ReturnMessage(http.StatusInternalServerError))
			c.Abort()
			return
		}

		if !verified {
			c.JSON(http.StatusUnauthorized, message.ReturnCustomMessage("User not allowed"))
			c.Abort()
			return
		}

		c.Next()
	}
}
