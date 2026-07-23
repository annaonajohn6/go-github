package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMiddleware_PanicStatusLogging(t *testing.T) {
	t.Run("Logger wraps Recovery", func(t *testing.T) {
		var logBuf bytes.Buffer
		logger := NewIOLogger(&logBuf)

		handler := LoggerMiddleware(logger)(
			RecoveryMiddleware()(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					panic("simulated handler panic")
				}),
			),
		)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected response status 500, got %d", rec.Code)
		}

		logStr := logBuf.String()
		if !strings.Contains(logStr, "500") {
			t.Errorf("Expected log output to contain status 500, got: %s", logStr)
		}
		if strings.Contains(logStr, "200") {
			t.Errorf("Log output incorrectly contains status 200: %s", logStr)
		}
	})

	t.Run("Recovery wraps Logger", func(t *testing.T) {
		var logBuf bytes.Buffer
		logger := NewIOLogger(&logBuf)

		handler := RecoveryMiddleware()(
			LoggerMiddleware(logger)(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					panic("simulated handler panic")
				}),
			),
		)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected response status 500, got %d", rec.Code)
		}

		logStr := logBuf.String()
		if !strings.Contains(logStr, "500") {
			t.Errorf("Expected log output to contain status 500, got: %s", logStr)
		}
		if strings.Contains(logStr, "200") {
			t.Errorf("Log output incorrectly contains status 200: %s", logStr)
		}
	})

	t.Run("Panic after headers written", func(t *testing.T) {
		var logBuf bytes.Buffer
		logger := NewIOLogger(&logBuf)

		handler := LoggerMiddleware(logger)(
			RecoveryMiddleware()(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusAccepted)
					w.Write([]byte("partial response"))
					panic("simulated handler panic after write")
				}),
			),
		)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		
		defer func() {
			if err := recover(); err != nil {
				t.Fatalf("Panic was not recovered by RecoveryMiddleware: %v", err)
			}
		}()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Errorf("Expected response status 202, got %d", rec.Code)
		}

		logStr := logBuf.String()
		if !strings.Contains(logStr, "202") {
			t.Errorf("Expected log output to contain status 202, got: %s", logStr)
		}
		if strings.Contains(logStr, "500") {
			t.Errorf("Log output incorrectly contains status 500: %s", logStr)
		}
	})
}
