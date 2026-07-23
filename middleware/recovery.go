package middleware

import (
	"net/http"
)

func RecoveryMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					if wr, ok := w.(WrapResponseWriter); ok {
						if !wr.WroteHeader() {
							wr.WriteHeader(http.StatusInternalServerError)
						}
					} else {
						w.WriteHeader(http.StatusInternalServerError)
					}
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
