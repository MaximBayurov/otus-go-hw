package internalhttp

import (
	"fmt"
	"net/http"
	"time"

	srvcontr "github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/server/contracts"
)

func loggingMiddleware(logger *srvcontr.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()

				lrw := newLoggingResponseWriter(w)
				next.ServeHTTP(lrw, r)

				duration := time.Since(start)

				ip := r.RemoteAddr
				if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
					ip = forwarded
				}

				timestamp := time.Now().Format("02/Jan/2006:15:04:05 -0700")
				method := r.Method
				path := r.URL.Path
				protocol := r.Proto
				statusCode := lrw.statusCode
				latency := duration.Milliseconds()

				userAgent := r.Header.Get("User-Agent")
				if userAgent == "" {
					userAgent = "-"
				}

				logLine := fmt.Sprintf(
					"%s [%s] %s %s %s %d %d \"%s\"",
					ip,
					timestamp,
					method,
					path,
					protocol,
					statusCode,
					latency,
					userAgent,
				)

				(*logger).Info(logLine)
			},
		)
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
