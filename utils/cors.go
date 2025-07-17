package utils

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func GetCorsConfig() gin.HandlerFunc {
	origins := os.Getenv("APP_ALLOWED_ORIGINS")
	originList := strings.Split(origins, ",")

	return cors.New(cors.Config{

		AllowOrigins: originList,
		AllowMethods: []string{"GET", "POST", "DELETE", "OPTIONS", "PUT"},
		AllowHeaders: []string{
			"authorization", "accept", "accept-encoding",
			"accept-language", "connection", "content-length",
			"content-type", "host", "origin", "referer", "user-agent",
		},
		AllowCredentials: true,
		MaxAge:           24 * time.Hour,
	})
}