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
			if isPublicRedirectRequest(r) {
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

func isPublicRedirectRequest(r *http.Request) bool {
	if r.Method != http.MethodGet {
		return false
	}

	trimmed := strings.Trim(strings.TrimSpace(r.URL.Path), "/")
	if trimmed == "" {
		return false
	}

	// /v1/{code}
	if strings.HasPrefix(trimmed, "v1/") {
		code := strings.TrimPrefix(trimmed, "v1/")
		return code != "" && !strings.Contains(code, "/")
	}

	// /{code}
	return !strings.Contains(trimmed, "/") && trimmed != "v1" && trimmed != "swagger"
}
