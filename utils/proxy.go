package utils

import (
	"bytes"
	"net/http"
)

type StreamingResponseWriter struct {
	Headers http.Header
	Body    *bytes.Buffer
	Status  int
}

func (w *StreamingResponseWriter) Header() http.Header {
	return w.Headers
}

func (w *StreamingResponseWriter) Write(b []byte) (int, error) {
	return w.Body.Write(b)
}

func (w *StreamingResponseWriter) WriteHeader(statusCode int) {
	w.Status = statusCode
}

func FlattenHeaders(h http.Header) map[string]string {
	result := make(map[string]string)
	for k, v := range h {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}