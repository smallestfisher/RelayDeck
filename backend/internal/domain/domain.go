package domain

import "time"

type EndpointType string

const (
	EndpointResponses       EndpointType = "responses"
	EndpointChatCompletions EndpointType = "chat_completions"
	EndpointEmbeddings      EndpointType = "embeddings"
)

type Capability string

const (
	CapabilityResponses Capability = "responses"
	CapabilityChat      Capability = "chat"
	CapabilityStreaming Capability = "streaming"
	CapabilityTools     Capability = "tools"
	CapabilityVision    Capability = "vision"
	CapabilityEmbedding Capability = "embedding"
)

type Scope string

const (
	ScopeResponses       Scope = "responses"
	ScopeChatCompletions Scope = "chat"
	ScopeEmbeddings      Scope = "embeddings"
	ScopeAdmin           Scope = "admin"
)

type APIKeyStatus string

const (
	APIKeyStatusActive  APIKeyStatus = "active"
	APIKeyStatusRevoked APIKeyStatus = "revoked"
	APIKeyStatusExpired APIKeyStatus = "expired"
)

type APIKey struct {
	ID              string
	UserID          string
	Prefix          string
	Hash            string
	Status          APIKeyStatus
	Scopes          []Scope
	AllowedModels   []string
	ExpiresAt       time.Time
	OwnerIsActive   bool
	RPM             int
	TPM             int
	AllowedCIDRs    []string
	MonthlyQuotaTPM int64
}

type GatewayPrincipal struct {
	APIKeyID string
	UserID   string
	Scopes   []Scope
}

type GatewayRequest struct {
	Endpoint             EndpointType
	Model                string
	Stream               bool
	RequiredCapabilities []Capability
}

type Model struct {
	ID           string
	Name         string
	Capabilities []Capability
}

type CircuitState string

const (
	CircuitClosed   CircuitState = "closed"
	CircuitOpen     CircuitState = "open"
	CircuitHalfOpen CircuitState = "half_open"
)

type UpstreamSite struct {
	ID           string
	Name         string
	BaseURL      string
	Credential   string
	Enabled      bool
	Weight       float64
	HealthScore  float64
	SuccessRate  float64
	LatencyMS    int
	Circuit      CircuitState
	Capabilities []Capability
}

type SiteModel struct {
	SiteID        string
	Model         string
	UpstreamModel string
	EndpointTypes []EndpointType
	Capabilities  []Capability
}

type RoutingPolicy struct {
	Mode               string
	MinimumHealthScore float64
	RetryCount         int
	CircuitCooldownSec int
}

type RouteCandidate struct {
	Site    UpstreamSite
	Mapping SiteModel
	Score   float64
}
