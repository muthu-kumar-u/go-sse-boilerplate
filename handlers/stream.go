package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	constants "github.com/muthu-kumar-u/go-sse/const"
	faceanalyze_events "github.com/muthu-kumar-u/go-sse/events/faceAnalyze"
	"github.com/muthu-kumar-u/go-sse/globals"
	appschema "github.com/muthu-kumar-u/go-sse/models"
	"github.com/muthu-kumar-u/go-sse/services"
	"github.com/muthu-kumar-u/go-sse/utils"
)

type StreamHandler struct {
	UserService    services.UserService
}

func NewFaceAnalyzeHandler(userService services.UserService) *StreamHandler {
	return &StreamHandler{
		UserService: userService,
	}	
}

func (h *StreamHandler) LogUserFace(c *gin.Context) {
	streamId := c.Query("stream")
	if streamId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "streamId is required"})
		return
	}

	sendEvent := func(event *appschema.EventMessage) {
        data, err := json.Marshal(event)
        if err != nil {
            log.Printf("marshal error: %v", err)
            return
        }
        payload := fmt.Sprintf("event: %s\ndata: %s\n\n", event.Event, data)
        globals.Stream.Publish(streamId, []byte(payload))
    }

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("multipart parse error: %v", err)
		sendEvent(&appschema.EventMessage{Code: 400, Message: "Invalid form data"})
		return
	}

	files := c.Request.MultipartForm.File["image"]
	if len(files) == 0 {
		sendEvent(&appschema.EventMessage{Code: 400, Message: "Missing image file"})
		return
	}

	fileHeader := files[0]
	file, err := fileHeader.Open()
	if err != nil {
		sendEvent(&appschema.EventMessage{Code: 400, Message: "Failed to open uploaded file"})
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !slices.Contains(constants.IMAGE_EXTENSIONS, ext) {
		sendEvent(&appschema.EventMessage{Code: 400, Message: "Only jpg, jpeg, png allowed"})
		return
	}

	sendEvent(&appschema.EventMessage{
		Code:       202,
		Event:      faceanalyze_events.EventProcessingImage,
		Message:    "Processing image",
		Completion: 25,
	})

	imageData, err := utils.PrepareImagePayloadFromBytes(file, fileHeader, constants.FACE_ANALYZE_PAYLOAD_FIELD_NAME)
	if err != nil {
		sendEvent(&appschema.EventMessage{Code: 500, Message: "Failed to process image"})
		return
	}

	sendEvent(&appschema.EventMessage{
		Code:       202,
		Event:      faceanalyze_events.EventAnalyzingFace,
		Message:    "Analyzing face",
		Completion: 50,
	})

	// Call FaceAnalyze API
	reqUrl := fmt.Sprintf("%s/%s", globals.FaceAnalyzeService.URL, constants.FACE_ANALYZE_SERVICE_PATHS[0])
	faceReq, err := http.NewRequest(http.MethodPost, reqUrl, imageData.MultipartBody)
	if err != nil {
		sendEvent(&appschema.EventMessage{Code: 500, Message: "Internal error"})
		return
	}
	faceReq.Header.Set("Authorization", os.Getenv("FACEANALYZE_SERVICE_AUTH_API_KEY"))
	faceReq.Header.Set("Content-Type", imageData.MultipartWriter.FormDataContentType())

	resp, err := globals.FaceAnalyzeService.Client.Do(faceReq)
	if err != nil {
		sendEvent(&appschema.EventMessage{Code: 500, Message: "Face analyze failed"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("FaceAnalyze failed: %s", string(body))
		sendEvent(&appschema.EventMessage{Code: 500, Message: "Face scan error"})
		return
	}

	var faResp appschema.FaceScannerResponse
	if err := utils.BindHttpResponseToStruct(resp, &faResp); err != nil {
		sendEvent(&appschema.EventMessage{Code: 500, Message: "Invalid face scan response"})
		return
	}

	sendEvent(&appschema.EventMessage{
		Code:       200,
		Event:      faceanalyze_events.EventCompleted,
		Data:       &appschema.FaceScanData{Quantitative: faResp.Data.Quantitative, Qualitative: faResp.Data.Qualitative},
		Message:    "Scan complete",
		Completion: 100,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "face scan complete",
		"stream":  streamId,
	})
}

func (h *StreamHandler) FaceLogStream(c *gin.Context) {
    streamId := c.Query("stream")
    if streamId == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "stream ID required"})
        return
    }

    // Set SSE headers
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")
    c.Header("Access-Control-Allow-Origin", "*")
    c.Header("X-Accel-Buffering", "no") // Important for some proxies

    // Get the flusher
    flusher, ok := c.Writer.(http.Flusher)
    if !ok {
        c.AbortWithStatus(http.StatusInternalServerError)
        log.Printf("[SSE] Stream %s: ResponseWriter doesn't support flushing", streamId)
        return
    }

    // Create a context that cancels when connection drops
    ctx, cancel := context.WithCancel(c.Request.Context())
    defer cancel()

    // Subscribe to the stream (note the receive-only channel)
    recvCh, _ := globals.Stream.Subscribe(ctx, streamId)
    defer func() {
        // Convert receive-only channel to bidirectional for unsubscribe
        if ch, ok := (interface{}(recvCh)).(chan []byte); ok {
            globals.Stream.Unsubscribe(streamId, ch)
        }
        log.Printf("[SSE] Stream %s: Unsubscribed", streamId)
    }()

    // Helper function to safely marshal JSON
    mustJSON := func(v interface{}) string {
        b, err := json.Marshal(v)
        if err != nil {
            log.Printf("JSON marshal error: %v", err)
            return "{}"
        }
        return string(b)
    }

    // Send initial handshake
    handshake := fmt.Sprintf("event: ready\ndata: %s\n\n", 
        mustJSON(map[string]interface{}{
            "code":      200,
            "stream_id": streamId,
            "ts":        time.Now().Unix(),
        }),
    )

    if _, err := c.Writer.Write([]byte(handshake)); err != nil {
        log.Printf("[SSE] Stream %s: Initial write failed: %v", streamId, err)
        return
    }
    flusher.Flush()

    log.Printf("[SSE] Stream %s: Connection established", streamId)

    // Heartbeat ticker
    heartbeat := time.NewTicker(15 * time.Second)
    defer heartbeat.Stop()

    // Main event loop
    for {
        select {
        case <-ctx.Done():
            log.Printf("[SSE] Stream %s: Context closed: %v", streamId, ctx.Err())
            return

        case <-heartbeat.C:
            // Send keep-alive comment
            if _, err := c.Writer.Write([]byte(": heartbeat\n\n")); err != nil {
                log.Printf("[SSE] Stream %s: Heartbeat failed: %v", streamId, err)
                return
            }
            flusher.Flush()

        case msg, ok := <-recvCh:
            if !ok {
                log.Printf("[SSE] Stream %s: Subscription channel closed", streamId)
                return
            }

            // Write the message
            if _, err := c.Writer.Write(msg); err != nil {
                log.Printf("[SSE] Stream %s: Write failed: %v", streamId, err)
                return
            }
            flusher.Flush()
        }
    }
}

func (h *StreamHandler) LogUserFaceLambda(ctx context.Context, req events.LambdaFunctionURLRequest) (*events.LambdaFunctionURLStreamingResponse, error) {
	streamId := uuid.NewString()
	go globals.Stream.CreateTemporaryStream(streamId, 2*time.Minute) // create temporary stream

	reader, writer := io.Pipe()
	done := make(chan struct{})

	go func() {
		defer func() {
			writer.Close()
			close(done)
			log.Printf("Stream completed or client disconnected for: %s", streamId)
		}()

		sendEvent := func(event *appschema.EventMessage) {
			data, err := json.Marshal(event)
			if err != nil {
				log.Printf("Marshal error: %v", err)
				return
			}
			payload := fmt.Sprintf("event: %s\ndata: %s\n\n", event.Event, data)
			writer.Write([]byte(payload))
			globals.Stream.Publish(streamId, []byte(payload))
		}

		sendEvent(&appschema.EventMessage{Code: 200,Event: "ready",Message: "Stream initialized", StreamID: streamId,Completion: 0,})
		sendEvent(&appschema.EventMessage{Code: 202,Event: faceanalyze_events.EventProcessingImage, Message: "Starting processing", Completion: 10})

		bodyBytes := []byte(req.Body)
		if req.IsBase64Encoded {
			decoded, err := base64.StdEncoding.DecodeString(req.Body)
			if err != nil {
				sendEvent(&appschema.EventMessage{Code: 400, Event: faceanalyze_events.EventError, Message: "Invalid base64"})
				return
			}
			bodyBytes = decoded
		}

		contentType := req.Headers["content-type"]
		if contentType == "" {
			contentType = req.Headers["Content-Type"]
		}
		boundary := extractBoundary(contentType)
		if boundary == "" {
			sendEvent(&appschema.EventMessage{Code: 400, Event: faceanalyze_events.EventError, Message: "Missing multipart boundary"})
			return
		}

		mr := multipart.NewReader(bytes.NewReader(bodyBytes), boundary)
		var fileData []byte
		var fileName string

		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				sendEvent(&appschema.EventMessage{Code: 400, Event: faceanalyze_events.EventError, Message: "Read error in multipart"})
				return
			}
			if part.FormName() == "image" {
				fileName = part.FileName()
				fileData, err = io.ReadAll(part)
				if err != nil {
					sendEvent(&appschema.EventMessage{Code: 400, Event: faceanalyze_events.EventError, Message: "Failed to read image data"})
					return
				}
				break
			}
		}

		if len(fileData) == 0 {
			sendEvent(&appschema.EventMessage{Code: 400, Event: faceanalyze_events.EventError, Message: "No image found"})
			return
		}

		ext := strings.ToLower(filepath.Ext(fileName))
		if !slices.Contains(constants.IMAGE_EXTENSIONS, ext) {
			sendEvent(&appschema.EventMessage{Code: 400, Event: faceanalyze_events.EventError, Message: "Unsupported file extension"})
			return
		}

		sendEvent(&appschema.EventMessage{Code: 202,Event: faceanalyze_events.EventAnalyzingFace,Message: "Analyzing face",Completion: 50})

		body := &bytes.Buffer{}
		mpWriter := multipart.NewWriter(body)
		part, err := mpWriter.CreateFormFile(constants.FACE_ANALYZE_PAYLOAD_FIELD_NAME, fileName)
		if err != nil {
			sendEvent(&appschema.EventMessage{Code: 500, Event: faceanalyze_events.EventError, Message: "Failed to prepare image for scan"})
			return
		}
		part.Write(fileData)
		mpWriter.Close()

		reqUrl := fmt.Sprintf("%s/%s", globals.FaceAnalyzeService.URL, constants.FACE_ANALYZE_SERVICE_PATHS[0])
		faceReq, err := http.NewRequest(http.MethodPost, reqUrl, body)
		if err != nil {
			sendEvent(&appschema.EventMessage{Code: 500, Event: faceanalyze_events.EventError, Message: "Request creation failed"})
			return
		}
		faceReq.Header.Set("Authorization", os.Getenv("FACEANALYZE_SERVICE_AUTH_API_KEY"))
		faceReq.Header.Set("Content-Type", mpWriter.FormDataContentType())

		resp, err := globals.FaceAnalyzeService.Client.Do(faceReq)
		if err != nil {
			sendEvent(&appschema.EventMessage{Code: 500, Event: faceanalyze_events.EventError, Message: "Face analyze call failed"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("Scan failed: %s", string(body))
			sendEvent(&appschema.EventMessage{Code: 500, Event: faceanalyze_events.EventError, Message: "Face scan error"})
			return
		}

		var faResp appschema.FaceScannerResponse
		if err := utils.BindHttpResponseToStruct(resp, &faResp); err != nil {
			sendEvent(&appschema.EventMessage{Code: 500, Event: faceanalyze_events.EventError, Message: "Failed to parse scan response"})
			return
		}

		sendEvent(&appschema.EventMessage{Code: 200,Event: faceanalyze_events.EventCompleted,Message: "Scan complete",Completion: 100,Data: &appschema.FaceScanData{Quantitative: faResp.Data.Quantitative, Qualitative: faResp.Data.Qualitative},})
	}()

	// heartbeat loop & client disconnect watch
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Printf("[⚠️] Client manually disconnected from stream: %s", streamId)
				return
			case <-done:
				return
			case <-ticker.C:
				writer.Write([]byte(": heartbeat\n\n"))
			}
		}
	}()

	return &events.LambdaFunctionURLStreamingResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type":                "text/event-stream",
			"Cache-Control":               "no-cache",
			"Connection":                  "keep-alive",
			"Access-Control-Allow-Origin": "*",
		},
		Body: reader,
	}, nil
}


func extractBoundary(contentType string) string {
    parts := strings.Split(contentType, ";")
    for _, part := range parts {
        part = strings.TrimSpace(part)
        if strings.HasPrefix(part, "boundary=") {
            return strings.Trim(strings.TrimPrefix(part, "boundary="), `"`)
        }
    }
    return ""
}