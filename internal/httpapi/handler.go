package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"url-shorter/internal/link"
)

type Handler struct {
	service *link.Service
	baseURL string
}

func NewHandler(service *link.Service, baseURL string) *Handler {
	return &Handler{service: service, baseURL: strings.TrimRight(baseURL, "/")}
}

func (h *Handler) getHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

func (h *Handler) getInfo(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, InfoResponse{Name: "URL Shortener", Version: "1.0.0"})
}

func (h *Handler) postLink(w http.ResponseWriter, r *http.Request) {
	var request LinkRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	created, err := h.service.Create(r.Context(), request.URL)
	if err != nil {
		switch {
		case errors.Is(err, link.ErrInvalidURL):
			http.Error(w, "invalid URL", http.StatusBadRequest)
		case errors.Is(err, link.ErrUnsupportedScheme):
			http.Error(w, "unsupported URL scheme", http.StatusBadRequest)
		case errors.Is(err, link.ErrMissingHostname):
			http.Error(w, "URL must contain a hostname", http.StatusBadRequest)
		default:
			log.Printf("failed to save link: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusCreated, LinkResponse{
		Code:        created.Code,
		ShortURL:    h.baseURL + "/" + created.Code,
		OriginalURL: created.OriginalURL,
	})
}

func (h *Handler) getLink(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Get(r.Context(), r.PathValue("code"))
	if err != nil {
		if errors.Is(err, link.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		log.Printf("failed to get link: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, item.OriginalURL, http.StatusFound)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
