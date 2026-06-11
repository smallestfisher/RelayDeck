package domain

import "time"

type EndpointType string

const (
	EndpointResponses       EndpointType = "responses"
	EndpointChatCompletions EndpointType = "chat_completions"
	EndpointEmbeddings      EndpointType = "embeddings"
)

type Protocol string

const (
	ProtocolOpenAIResponses   Protocol = "openai_responses"
	ProtocolOpenAIChat        Protocol = "openai_chat_completions"
	ProtocolAnthropicMessages Protocol = "anthropic_messages"
	ProtocolGeminiGenerate    Protocol = "gemini_generate_content"
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

type UserRole string

const (
	UserRoleOwner     UserRole = "owner"
	UserRoleAdmin     UserRole = "admin"
	UserRoleDeveloper UserRole = "developer"
	UserRoleViewer    UserRole = "viewer"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusBlocked  UserStatus = "blocked"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Role         UserRole
	Status       UserStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

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

type UpstreamPlatformKind string

const (
	PlatformKindNewAPI  UpstreamPlatformKind = "new_api"
	PlatformKindSub2API UpstreamPlatformKind = "sub2api"
)

type UpstreamCredentialKind string

const (
	CredentialKindNone         UpstreamCredentialKind = "none"
	CredentialKindCookie       UpstreamCredentialKind = "cookie"
	CredentialKindAccessToken  UpstreamCredentialKind = "access_token"
	CredentialKindRefreshToken UpstreamCredentialKind = "refresh_token"
	CredentialKindJSON         UpstreamCredentialKind = "json"
)

type UpstreamAPIStatus string

const (
	UpstreamAPIStatusUnknown  UpstreamAPIStatus = "unknown"
	UpstreamAPIStatusHealthy  UpstreamAPIStatus = "healthy"
	UpstreamAPIStatusWarning  UpstreamAPIStatus = "warning"
	UpstreamAPIStatusFailed   UpstreamAPIStatus = "failed"
	UpstreamAPIStatusDisabled UpstreamAPIStatus = "disabled"
)

type AccountCredentialStatus string

const (
	AccountCredentialStatusNotConfigured  AccountCredentialStatus = "not_configured"
	AccountCredentialStatusValid          AccountCredentialStatus = "valid"
	AccountCredentialStatusExpired        AccountCredentialStatus = "expired"
	AccountCredentialStatusFailed         AccountCredentialStatus = "failed"
	AccountCredentialStatusActionRequired AccountCredentialStatus = "action_required"
)

type UpstreamCheckinStatus string

const (
	CheckinStatusUnsupported    UpstreamCheckinStatus = "unsupported"
	CheckinStatusNotConfigured  UpstreamCheckinStatus = "not_configured"
	CheckinStatusChecked        UpstreamCheckinStatus = "checked"
	CheckinStatusUnchecked      UpstreamCheckinStatus = "unchecked"
	CheckinStatusFailed         UpstreamCheckinStatus = "failed"
	CheckinStatusActionRequired UpstreamCheckinStatus = "action_required"
)

type UpstreamErrorClass string

const (
	UpstreamErrorAuthError         UpstreamErrorClass = "auth_error"
	UpstreamErrorCredentialMissing UpstreamErrorClass = "credential_missing"
	UpstreamErrorCredentialExpired UpstreamErrorClass = "credential_expired"
	UpstreamErrorQuota             UpstreamErrorClass = "quota_error"
	UpstreamErrorRateLimit         UpstreamErrorClass = "rate_limit"
	UpstreamErrorTimeout           UpstreamErrorClass = "timeout"
	UpstreamErrorTransport         UpstreamErrorClass = "transport_error"
	UpstreamErrorProtocolMismatch  UpstreamErrorClass = "protocol_mismatch"
	UpstreamErrorInvalidResponse   UpstreamErrorClass = "invalid_response"
	UpstreamErrorUpstream5xx       UpstreamErrorClass = "upstream_5xx"
	UpstreamErrorUnsupported       UpstreamErrorClass = "unsupported"
	UpstreamErrorActionRequired    UpstreamErrorClass = "action_required"
	UpstreamErrorUnknown           UpstreamErrorClass = "unknown_error"
)

type UpstreamAccount struct {
	ID                         string
	Name                       string
	Code                       string
	PlatformKind               UpstreamPlatformKind
	BaseURL                    string
	Enabled                    bool
	IncludeInRouting           bool
	Priority                   int
	APIKeyEncrypted            string
	APIKeyPrefix               string
	AccountCredentialKind      UpstreamCredentialKind
	AccountCredentialEncrypted string
	AutoSyncModels             bool
	AutoRefreshQuota           bool
	AutoCheckin                bool
	Note                       string
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
}

func (a UpstreamAccount) HasAPICredential() bool {
	return a.APIKeyPrefix != "" || a.APIKeyEncrypted != ""
}

func (a UpstreamAccount) HasAccountCredential() bool {
	return a.AccountCredentialKind != "" && a.AccountCredentialKind != CredentialKindNone && a.AccountCredentialEncrypted != ""
}

type UpstreamAccountStatus struct {
	UpstreamAccountID    string
	APIStatus            UpstreamAPIStatus
	AccountStatus        AccountCredentialStatus
	CheckinStatus        UpstreamCheckinStatus
	ModelCount           int
	LatencyMS            int
	BalanceAmount        float64
	BalanceUnit          string
	LastAPICheckedAt     time.Time
	LastAccountCheckedAt time.Time
	LastModelSyncedAt    time.Time
	LastCheckinAt        time.Time
	LastErrorClass       UpstreamErrorClass
	LastErrorMessage     string
	ActionRequiredReason string
	UpdatedAt            time.Time
}

func (s UpstreamAccountStatus) CanRouteTraffic() bool {
	return s.APIStatus == UpstreamAPIStatusHealthy || s.APIStatus == UpstreamAPIStatusWarning
}

func (s UpstreamAccountStatus) NeedsManualAction() bool {
	return s.AccountStatus == AccountCredentialStatusActionRequired || s.CheckinStatus == CheckinStatusActionRequired
}

type UpstreamSyncedModel struct {
	ID                     string
	UpstreamAccountID      string
	NormalizedModelName    string
	UpstreamModelName      string
	DisplayName            string
	NativeWireProtocol     Protocol
	SupportedWireProtocols []Protocol
	Capabilities           []Capability
	Status                 string
	RawMetadata            map[string]any
	LastSyncedAt           time.Time
}

type UpstreamAccountEvent struct {
	ID                string
	UpstreamAccountID string
	Operation         string
	Status            string
	ErrorClass        UpstreamErrorClass
	Message           string
	LatencyMS         int
	Metadata          map[string]any
	CreatedAt         time.Time
}

type ModelSyncResult struct {
	AccountID     string
	CreatedModels int
	UpdatedModels int
	Models        []UpstreamSyncedModel
}

type QuotaRefreshResult struct {
	Status        UpstreamAccountStatus
	BalanceAmount float64
	BalanceUnit   string
}

type CheckinResult struct {
	Status  UpstreamCheckinStatus
	Message string
}
