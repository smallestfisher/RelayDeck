package admin

import (
	"encoding/json"
	"net/http"

	"github.com/smallestfisher/relaydeck/backend/internal/domain"
)

type Store interface {
	APIKeys() []domain.APIKey
	Models() []domain.Model
	Sites() []domain.UpstreamSite
}

type Handler struct {
	store Store
}

func New(store Store) http.Handler {
	h := &Handler{store: store}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/admin/summary", h.handleSummary)
	return mux
}

func (h *Handler) handleSummary(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]int{
		"sites":    len(h.store.Sites()),
		"models":   len(h.store.Models()),
		"api_keys": len(h.store.APIKeys()),
	})
}
