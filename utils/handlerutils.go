package utils

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/gin-gonic/gin"
	appschema "github.com/muthu-kumar-u/go-sse/models"
)

func GetUserIdFromHeader(c *gin.Context) (string, error) {
	userId, exists := c.Get("id")
	if !exists {
		return "", fmt.Errorf("userId is not exist")
	}

	return userId.(string), nil
}

func PrepareImagePayloadFromBytes(file multipart.File, header *multipart.FileHeader, fieldName string) (*appschema.ImagePayload, error) {
    rawBytes, err := io.ReadAll(file)
    if err != nil {
        return nil, fmt.Errorf("failed to read image file: %w", err)
    }

    multipartBuf := &bytes.Buffer{}
    writer := multipart.NewWriter(multipartBuf)
    
    part, err := writer.CreateFormFile(fieldName, header.Filename)
    if err != nil {
        return nil, fmt.Errorf("failed to create multipart form file: %w", err)
    }
    
    if _, err := part.Write(rawBytes); err != nil {
        return nil, fmt.Errorf("failed to write to multipart form: %w", err)
    }
    
    if err := writer.Close(); err != nil {
        return nil, fmt.Errorf("failed to close multipart writer: %w", err)
    }

    return &appschema.ImagePayload{
        RawBytes:        rawBytes,
        MultipartBody:   multipartBuf,
        MultipartWriter: writer,
    }, nil
}
