package httpapi

type HealthResponse struct {
	Status string `json:"status"`
}

type InfoResponse struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type LinkRequest struct {
	URL string `json:"url"`
}

type LinkResponse struct {
	Code        string `json:"code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
