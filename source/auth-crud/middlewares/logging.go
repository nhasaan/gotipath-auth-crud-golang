package middlewares

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"auth-crud/loggers"
	"auth-crud/utils"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
	buf    bytes.Buffer
}

func (rr *responseRecorder) WriteHeader(statusCode int) {
	rr.status = statusCode
	rr.ResponseWriter.WriteHeader(statusCode)
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	rr.buf.Write(b)
	return rr.ResponseWriter.Write(b)
}

// Logging wraps handlers to log request/response with headers/body and errors as JSON.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Read & restore body for logging
		var reqBody []byte
		if r.Body != nil {
			reqBody, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		recorder := &responseRecorder{ResponseWriter: w, status: 200}
		reqID := utils.GetOrSetRequestID(w, r)

		next.ServeHTTP(recorder, r)

		dur := time.Since(start)
		fields := map[string]interface{}{
			"request_id":  reqID,
			"level":       "info",
			"method":      r.Method,
			"path":        r.URL.Path,
			"status":      recorder.status,
			"duration_ms": dur.Milliseconds(),
			"req_headers": r.Header,
			"req_body":    string(reqBody),
			"resp_body":   recorder.buf.String(),
		}
		if recorder.status >= 400 {
			fields["level"] = "error"
		}
		loggers.Log(fields)
	})
}
