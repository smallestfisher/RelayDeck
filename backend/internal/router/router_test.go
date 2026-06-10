package router

import (
	"testing"

	"github.com/smallestfisher/relaydeck/backend/internal/domain"
)

func TestSelectCandidateFiltersByCompatibility(t *testing.T) {
	req := domain.GatewayRequest{
		Endpoint:             domain.EndpointChatCompletions,
		Model:                "gpt-4o-mini",
		Stream:               true,
		RequiredCapabilities: []domain.Capability{domain.CapabilityChat, domain.CapabilityStreaming},
	}
	principal := domain.GatewayPrincipal{APIKeyID: "key_1", UserID: "user_1"}
	policy := domain.RoutingPolicy{Mode: "smart", MinimumHealthScore: 70}
	sites := []domain.UpstreamSite{
		{ID: "disabled", Enabled: false, HealthScore: 100, SuccessRate: 100, LatencyMS: 20, Circuit: domain.CircuitClosed, Capabilities: []domain.Capability{domain.CapabilityChat, domain.CapabilityStreaming}},
		{ID: "circuit", Enabled: true, HealthScore: 100, SuccessRate: 100, LatencyMS: 20, Circuit: domain.CircuitOpen, Capabilities: []domain.Capability{domain.CapabilityChat, domain.CapabilityStreaming}},
		{ID: "low-health", Enabled: true, HealthScore: 20, SuccessRate: 100, LatencyMS: 20, Circuit: domain.CircuitClosed, Capabilities: []domain.Capability{domain.CapabilityChat, domain.CapabilityStreaming}},
		{ID: "no-stream", Enabled: true, HealthScore: 100, SuccessRate: 100, LatencyMS: 20, Circuit: domain.CircuitClosed, Capabilities: []domain.Capability{domain.CapabilityChat}},
		{ID: "winner", Enabled: true, HealthScore: 95, SuccessRate: 98, LatencyMS: 120, Weight: 80, Circuit: domain.CircuitClosed, Capabilities: []domain.Capability{domain.CapabilityChat, domain.CapabilityStreaming}},
	}
	mappings := []domain.SiteModel{
		{SiteID: "disabled", Model: "gpt-4o-mini", UpstreamModel: "gpt-4o-mini", EndpointTypes: []domain.EndpointType{domain.EndpointChatCompletions}, Capabilities: []domain.Capability{domain.CapabilityChat, domain.CapabilityStreaming}},
		{SiteID: "circuit", Model: "gpt-4o-mini", UpstreamModel: "gpt-4o-mini", EndpointTypes: []domain.EndpointType{domain.EndpointChatCompletions}, Capabilities: []domain.Capability{domain.CapabilityChat, domain.CapabilityStreaming}},
		{SiteID: "low-health", Model: "gpt-4o-mini", UpstreamModel: "gpt-4o-mini", EndpointTypes: []domain.EndpointType{domain.EndpointChatCompletions}, Capabilities: []domain.Capability{domain.CapabilityChat, domain.CapabilityStreaming}},
		{SiteID: "no-stream", Model: "gpt-4o-mini", UpstreamModel: "gpt-4o-mini", EndpointTypes: []domain.EndpointType{domain.EndpointChatCompletions}, Capabilities: []domain.Capability{domain.CapabilityChat}},
		{SiteID: "winner", Model: "gpt-4o-mini", UpstreamModel: "gpt-4o-mini", EndpointTypes: []domain.EndpointType{domain.EndpointChatCompletions}, Capabilities: []domain.Capability{domain.CapabilityChat, domain.CapabilityStreaming}},
	}

	candidate, err := SelectCandidate(req, principal, sites, mappings, policy)
	if err != nil {
		t.Fatalf("expected candidate, got error: %v", err)
	}
	if candidate.Site.ID != "winner" {
		t.Fatalf("expected winner, got %s", candidate.Site.ID)
	}
}

func TestSelectCandidateUsesSmartScore(t *testing.T) {
	req := domain.GatewayRequest{Endpoint: domain.EndpointChatCompletions, Model: "gpt-4o-mini", RequiredCapabilities: []domain.Capability{domain.CapabilityChat}}
	principal := domain.GatewayPrincipal{APIKeyID: "key_1", UserID: "user_1"}
	policy := domain.RoutingPolicy{Mode: "smart", MinimumHealthScore: 50}
	sites := []domain.UpstreamSite{
		{ID: "slow", Enabled: true, HealthScore: 82, SuccessRate: 90, LatencyMS: 900, Weight: 80, Circuit: domain.CircuitClosed, Capabilities: []domain.Capability{domain.CapabilityChat}},
		{ID: "fast", Enabled: true, HealthScore: 96, SuccessRate: 99, LatencyMS: 120, Weight: 50, Circuit: domain.CircuitClosed, Capabilities: []domain.Capability{domain.CapabilityChat}},
	}
	mappings := []domain.SiteModel{
		{SiteID: "slow", Model: "gpt-4o-mini", UpstreamModel: "gpt-4o-mini", EndpointTypes: []domain.EndpointType{domain.EndpointChatCompletions}, Capabilities: []domain.Capability{domain.CapabilityChat}},
		{SiteID: "fast", Model: "gpt-4o-mini", UpstreamModel: "gpt-4o-mini", EndpointTypes: []domain.EndpointType{domain.EndpointChatCompletions}, Capabilities: []domain.Capability{domain.CapabilityChat}},
	}

	candidate, err := SelectCandidate(req, principal, sites, mappings, policy)
	if err != nil {
		t.Fatalf("expected candidate, got error: %v", err)
	}
	if candidate.Site.ID != "fast" {
		t.Fatalf("expected fast candidate, got %s with score %.2f", candidate.Site.ID, candidate.Score)
	}
}

func TestSelectCandidateReturnsErrorWhenNoneMatch(t *testing.T) {
	_, err := SelectCandidate(
		domain.GatewayRequest{Endpoint: domain.EndpointResponses, Model: "gpt-4o-mini", RequiredCapabilities: []domain.Capability{domain.CapabilityResponses}},
		domain.GatewayPrincipal{APIKeyID: "key_1", UserID: "user_1"},
		[]domain.UpstreamSite{{ID: "chat-only", Enabled: true, HealthScore: 100, Circuit: domain.CircuitClosed, Capabilities: []domain.Capability{domain.CapabilityChat}}},
		[]domain.SiteModel{{SiteID: "chat-only", Model: "gpt-4o-mini", EndpointTypes: []domain.EndpointType{domain.EndpointChatCompletions}, Capabilities: []domain.Capability{domain.CapabilityChat}}},
		domain.RoutingPolicy{Mode: "smart", MinimumHealthScore: 50},
	)
	if err == nil {
		t.Fatal("expected no candidate error")
	}
}
