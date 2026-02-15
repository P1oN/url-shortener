package httpapi

import "net/http"

// PostV1Shorten satisfies the generated OpenAPI server interface.
func (h *Handlers) PostV1Shorten(w http.ResponseWriter, r *http.Request) {
	h.CreateShortURLHandler(w, r)
}
