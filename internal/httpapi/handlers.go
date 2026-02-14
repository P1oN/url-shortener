package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"url-shortener-go/internal/models"
	"url-shortener-go/internal/service"

	"github.com/gorilla/mux"
)

const maxBodyBytes = 1 << 20

type Handlers struct {
	service *service.Service
}

func NewHandlers(service *service.Service) *Handlers {
	return &Handlers{service: service}
}

type createShortURLRequest struct {
	OriginalURL      string `json:"original_url"`
	CustomCode       string `json:"custom_code,omitempty"`
	ExpiresInSeconds *int64 `json:"expires_in_seconds,omitempty"`
}

type createShortURLResponse struct {
	ShortURL  string `json:"short_url"`
	Code      string `json:"code"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

func (h *Handlers) CreateShortURLHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var payload createShortURLRequest
	if err := decoder.Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if payload.OriginalURL == "" {
		writeError(w, http.StatusBadRequest, "missing_url", "original_url is required")
		return
	}

	var expiresIn time.Duration
	if payload.ExpiresInSeconds != nil {
		if *payload.ExpiresInSeconds < 0 {
			writeError(w, http.StatusBadRequest, "invalid_expires_in", "expires_in_seconds must be positive")
			return
		}
		expiresIn = time.Duration(*payload.ExpiresInSeconds) * time.Second
	}

	url, err := h.service.CreateShortURL(r.Context(), models.CreateURLOptions{
		OriginalURL: payload.OriginalURL,
		CustomCode:  payload.CustomCode,
		ExpiresIn:   expiresIn,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidURL):
			writeError(w, http.StatusBadRequest, "invalid_url", "invalid URL")
		case errors.Is(err, service.ErrConflict):
			writeError(w, http.StatusConflict, "short_code_conflict", "short code already exists")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "unexpected error")
		}
		return
	}

	response := createShortURLResponse{
		ShortURL: h.service.GenerateShortURL(url.ShortCode),
		Code:     url.ShortCode,
	}
	if url.ExpiresAt != nil {
		response.ExpiresAt = url.ExpiresAt.UTC().Format(time.RFC3339)
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *Handlers) GetFullURLHandler(w http.ResponseWriter, r *http.Request) {
	shortCode := mux.Vars(r)["code"]
	if shortCode == "" {
		writeError(w, http.StatusBadRequest, "missing_code", "missing short code")
		return
	}

	url, err := h.service.GetFullURL(r.Context(), shortCode)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "URL not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "unexpected error")
		return
	}

	http.Redirect(w, r, url.OriginalURL, http.StatusFound)
}

func (h *Handlers) HealthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
