package middleware

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type Logger interface {
	Log(status int, method, path string, duration time.Duration)
}

type ioLogger struct {
	w io.Writer
}

func NewIOLogger(w io.Writer) Logger {
	return &ioLogger{w: w}
}

func (l *ioLogger) Log(status int, method, path string, duration time.Duration) {
	fmt.Fprintf(l.w, "[HTTP] %d | %s %s | %v\n", status, method, path, duration)
}

func LoggerMiddleware(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := NewWrapResponseWriter(w)
			start := time.Now()
			
			panicked := true
			defer func() {
				duration := time.Since(start)
				status := ww.Status()
				if panicked && !ww.WroteHeader() {
					status = http.StatusInternalServerError
				}
				logger.Log(status, r.Method, r.URL.Path, duration)
			}()
			
			next.ServeHTTP(ww, r)
			panicked = false
			
		})
	}
}
