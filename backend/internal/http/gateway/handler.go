package gateway

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/auth"
	"github.com/smallestfisher/relaydeck/backend/internal/domain"
	deckrouter "github.com/smallestfisher/relaydeck/backend/internal/router"
	"github.com/smallestfisher/relaydeck/backend/internal/upstream"
)

type Store interface {
	APIKeys() []domain.APIKey
	Models() []domain.Model
	Sites() []domain.UpstreamSite
	Mappings() []domain.SiteModel
}

type NowFunc func() time.Time

type Handler struct {
	store    Store
	upstream *upstream.Client
	now      NowFunc
}

func New(store Store, upstreamClient *upstream.Client, now NowFunc) http.Handler {
	h := &Handler{store: store, upstream: upstreamClient, now: now}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/models", h.handleModels)
	mux.HandleFunc("POST /v1/chat/completions", h.handleChatCompletions)
	mux.HandleFunc("POST /v1/responses", h.handleResponses)
	return mux
}

func (h *Handler) handleModels(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.authenticate(w, r, domain.GatewayRequest{Endpoint: domain.EndpointChatCompletions, Model: "gpt-4o-mini"}); !ok {
		return
	}
	items := make([]map[string]any, 0, len(h.store.Models()))
	for _, model := range h.store.Models() {
		items = append(items, map[string]any{"id": model.ID, "object": "model", "owned_by": "relaydeck"})
	}
	writeJSON(w, http.StatusOK, map[string]any{"object": "list", "data": items})
}

func (h *Handler) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	body, payload, ok := readPayload(w, r)
	if !ok {
		return
	}
	model, _ := payload["model"].(string)
	stream, _ := payload["stream"].(bool)
	req := domain.GatewayRequest{
		Endpoint:             domain.EndpointChatCompletions,
		Model:                model,
		Stream:               stream,
		RequiredCapabilities: []domain.Capability{domain.CapabilityChat},
	}
	if stream {
		req.RequiredCapabilities = append(req.RequiredCapabilities, domain.CapabilityStreaming)
	}
	principal, ok := h.authenticate(w, r, req)
	if !ok {
		return
	}
	candidate, err := deckrouter.SelectCandidate(req, principal, h.store.Sites(), h.store.Mappings(), domain.RoutingPolicy{Mode: "smart", MinimumHealthScore: 1})
	if err != nil {
		writeOpenAIError(w, http.StatusBadGateway, "no_upstream_available", err.Error())
		return
	}
	payload["model"] = candidate.Mapping.UpstreamModel
	mappedBody, err := json.Marshal(payload)
	if err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "invalid_request", "request body cannot be encoded")
		return
	}
	if len(mappedBody) == 0 {
		mappedBody = body
	}
	resp, err := h.upstream.DoJSON(r.Context(), candidate.Site, "/v1/chat/completions", mappedBody)
	if err != nil {
		status := http.StatusBadGateway
		var upstreamErr *upstream.Error
		if errors.As(err, &upstreamErr) {
			status = upstreamErr.StatusCode
		}
		writeOpenAIError(w, status, "upstream_error", err.Error())
		return
	}
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(resp.Body)
}

func (h *Handler) handleResponses(w http.ResponseWriter, r *http.Request) {
	_, payload, ok := readPayload(w, r)
	if !ok {
		return
	}
	model, _ := payload["model"].(string)
	req := domain.GatewayRequest{Endpoint: domain.EndpointResponses, Model: model, RequiredCapabilities: []domain.Capability{domain.CapabilityResponses}}
	principal, ok := h.authenticate(w, r, req)
	if !ok {
		return
	}
	_, err := deckrouter.SelectCandidate(req, principal, h.store.Sites(), h.store.Mappings(), domain.RoutingPolicy{Mode: "smart", MinimumHealthScore: 1})
	if err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "unsupported_endpoint", "no upstream supports /v1/responses for this model")
		return
	}
	writeOpenAIError(w, http.StatusNotImplemented, "not_implemented", "responses passthrough is not enabled in this slice")
}

func (h *Handler) authenticate(w http.ResponseWriter, r *http.Request, req domain.GatewayRequest) (domain.GatewayPrincipal, bool) {
	secret := bearerSecret(r.Header.Get("Authorization"))
	if secret == "" {
		writeOpenAIError(w, http.StatusUnauthorized, "unauthorized", "missing bearer token")
		return domain.GatewayPrincipal{}, false
	}
	for _, key := range h.store.APIKeys() {
		principal, err := auth.VerifyGatewayKey(secret, key, req, h.now())
		if err == nil {
			return principal, true
		}
	}
	writeOpenAIError(w, http.StatusUnauthorized, "unauthorized", "invalid bearer token")
	return domain.GatewayPrincipal{}, false
}

func readPayload(w http.ResponseWriter, r *http.Request) ([]byte, map[string]any, bool) {
	body := new(bytes.Buffer)
	if _, err := body.ReadFrom(r.Body); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "invalid_request", "request body cannot be read")
		return nil, nil, false
	}
	payload := map[string]any{}
	if err := json.Unmarshal(body.Bytes(), &payload); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "invalid_json", "request body must be JSON")
		return nil, nil, false
	}
	return body.Bytes(), payload, true
}

func bearerSecret(header string) string {
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeOpenAIError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]any{"message": message, "type": "relaydeck_error", "code": code}})
}

func copyHeader(dst http.Header, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}
