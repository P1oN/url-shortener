package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

func AuthMiddleware(apiKey string, allowUnauthedDocs bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v1/health" {
				next.ServeHTTP(w, r)
				return
			}
			if allowUnauthedDocs && (r.URL.Path == "/swagger" || strings.HasPrefix(r.URL.Path, "/swagger/")) {
				next.ServeHTTP(w, r)
				return
			}

			token := strings.TrimSpace(r.Header.Get("Authorization"))
			if token == "" || !strings.HasPrefix(token, "Bearer ") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			provided := strings.TrimPrefix(token, "Bearer ")
			if subtle.ConstantTimeCompare([]byte(provided), []byte(apiKey)) != 1 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
