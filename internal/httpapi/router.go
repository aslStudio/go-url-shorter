package httpapi

import "net/http"

func NewRouter(handler *Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.getHealth)
	mux.HandleFunc("GET /api/info", handler.getInfo)
	mux.HandleFunc("POST /api/link", handler.postLink)
	mux.HandleFunc("GET /{code}", handler.getLink)
	return mux
}
