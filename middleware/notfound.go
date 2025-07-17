package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nelsonin-research-org/clenz-stream/message"
)

func PathNotFound() gin.HandlerFunc {
	return func (c *gin.Context)  {
		c.JSON(http.StatusNotFound, message.ReturnCustomMessage("path not found"))
		c.Abort()
		return
	}
}