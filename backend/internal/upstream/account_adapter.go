package upstream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/domain"
)

type AccountAdapter interface {
	TestAPI(ctx context.Context, account domain.UpstreamAccount, apiKey string) (domain.UpstreamAccountStatus, error)
	TestAccountCredential(ctx context.Context, account domain.UpstreamAccount, credential string) (domain.AccountCredentialTestResult, error)
	SyncModels(ctx context.Context, account domain.UpstreamAccount, apiKey string) (domain.ModelSyncResult, domain.UpstreamAccountStatus, error)
	RefreshQuota(ctx context.Context, account domain.UpstreamAccount, apiKey string, accountCredential string) (domain.QuotaRefreshResult, error)
	Checkin(ctx context.Context, account domain.UpstreamAccount, accountCredential string) (domain.CheckinResult, error)
	TestCall(ctx context.Context, account domain.UpstreamAccount, apiKey string, modelName string, protocol string, streaming bool, message string) (domain.UpstreamTestCallResult, error)
}

type AccountAdapterRegistry struct {
	adapters map[domain.UpstreamPlatformKind]AccountAdapter
}

func NewAccountAdapterRegistry(adapters map[domain.UpstreamPlatformKind]AccountAdapter) AccountAdapterRegistry {
	copied := make(map[domain.UpstreamPlatformKind]AccountAdapter, len(adapters))
	for kind, adapter := range adapters {
		copied[kind] = adapter
	}
	return AccountAdapterRegistry{adapters: copied}
}

func DefaultAccountAdapterRegistry(httpClient *http.Client, timeout time.Duration) AccountAdapterRegistry {
	return NewAccountAdapterRegistry(map[domain.UpstreamPlatformKind]AccountAdapter{
		domain.PlatformKindNewAPI:  NewNewAPIAccountAdapter(httpClient, timeout),
		domain.PlatformKindSub2API: NewSub2APIAccountAdapter(httpClient, timeout),
	})
}

func (r AccountAdapterRegistry) For(kind domain.UpstreamPlatformKind) (AccountAdapter, bool) {
	adapter, ok := r.adapters[kind]
	return adapter, ok
}

type accountHTTPAdapter struct {
	httpClient *http.Client
	timeout    time.Duration
}

func newAccountHTTPAdapter(httpClient *http.Client, timeout time.Duration) accountHTTPAdapter {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}
	if timeout > 0 && httpClient.Timeout == 0 {
		httpClient.Timeout = timeout
	}
	return accountHTTPAdapter{httpClient: httpClient, timeout: timeout}
}

type adapterResponse struct {
	statusCode int
	body       []byte
}

type newAPIAccessCredential struct {
	AccessToken string `json:"access_token"`
	UserID      string `json:"user_id"`
}

type sub2APIRefreshCredential struct {
	RefreshToken string `json:"refresh_token"`
}

func (a accountHTTPAdapter) do(ctx context.Context, method string, account domain.UpstreamAccount, path string, authHeader string, body io.Reader) (adapterResponse, error) {
	if strings.TrimSpace(account.BaseURL) == "" {
		return adapterResponse{}, errors.New("upstream base url is required")
	}
	ctx, cancel := context.WithTimeout(ctx, requestTimeout(a.timeout))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(account.BaseURL, "/")+path, body)
	if err != nil {
		return adapterResponse{}, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return adapterResponse{}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return adapterResponse{}, err
	}
	return adapterResponse{statusCode: resp.StatusCode, body: respBody}, nil
}

func (a accountHTTPAdapter) doNewAPIUser(ctx context.Context, method string, account domain.UpstreamAccount, path string, credential newAPIAccessCredential, body io.Reader) (adapterResponse, error) {
	if strings.TrimSpace(account.BaseURL) == "" {
		return adapterResponse{}, errors.New("upstream base url is required")
	}
	ctx, cancel := context.WithTimeout(ctx, requestTimeout(a.timeout))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(account.BaseURL, "/")+path, body)
	if err != nil {
		return adapterResponse{}, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", credential.AccessToken)
	req.Header.Set("New-Api-User", credential.UserID)
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return adapterResponse{}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return adapterResponse{}, err
	}
	return adapterResponse{statusCode: resp.StatusCode, body: respBody}, nil
}

func (a accountHTTPAdapter) doCookie(ctx context.Context, method string, account domain.UpstreamAccount, path string, credential string, body io.Reader) (adapterResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout(a.timeout))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(account.BaseURL, "/")+path, body)
	if err != nil {
		return adapterResponse{}, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	applyAccountCredential(req, credential)
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return adapterResponse{}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return adapterResponse{}, err
	}
	return adapterResponse{statusCode: resp.StatusCode, body: respBody}, nil
}

func requestTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return 10 * time.Second
	}
	return timeout
}

func applyAccountCredential(req *http.Request, credential string) {
	credential = strings.TrimSpace(credential)
	if credential == "" {
		return
	}
	if strings.HasPrefix(strings.ToLower(credential), "bearer ") {
		req.Header.Set("Authorization", credential)
		return
	}
	if strings.Contains(credential, "=") || strings.Contains(strings.ToLower(credential), "session") {
		req.Header.Set("Cookie", credential)
		return
	}
	req.Header.Set("Authorization", "Bearer "+credential)
}

func parseNewAPIAccessCredential(credential string) (newAPIAccessCredential, error) {
	credential = strings.TrimSpace(credential)
	if credential == "" {
		return newAPIAccessCredential{}, errors.New("account credential is not configured")
	}
	payload := map[string]any{}
	decoder := json.NewDecoder(strings.NewReader(credential))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		return newAPIAccessCredential{}, errors.New("new-api account credential must be JSON with access_token and user_id")
	}
	accessToken, _ := payload["access_token"].(string)
	userID, ok := userIDString(payload["user_id"])
	if strings.TrimSpace(accessToken) == "" || !ok || strings.TrimSpace(userID) == "" {
		return newAPIAccessCredential{}, errors.New("new-api account credential requires access_token and user_id")
	}
	return newAPIAccessCredential{AccessToken: strings.TrimSpace(accessToken), UserID: strings.TrimSpace(userID)}, nil
}

func parseSub2APIRefreshCredential(credential string) (sub2APIRefreshCredential, error) {
	credential = strings.TrimSpace(credential)
	if credential == "" {
		return sub2APIRefreshCredential{}, errors.New("account credential is not configured")
	}
	if strings.HasPrefix(credential, "rt_") {
		return sub2APIRefreshCredential{RefreshToken: credential}, nil
	}
	payload := map[string]any{}
	if err := json.Unmarshal([]byte(credential), &payload); err != nil {
		return sub2APIRefreshCredential{}, errors.New("sub2api account credential must be JSON with refresh_token")
	}
	refreshToken, _ := payload["refresh_token"].(string)
	refreshToken = strings.TrimSpace(refreshToken)
	if !strings.HasPrefix(refreshToken, "rt_") {
		return sub2APIRefreshCredential{}, errors.New("sub2api refresh token must start with rt_")
	}
	return sub2APIRefreshCredential{RefreshToken: refreshToken}, nil
}

func encodeSub2APIRefreshCredential(refreshToken string) string {
	payload, err := json.Marshal(sub2APIRefreshCredential{RefreshToken: strings.TrimSpace(refreshToken)})
	if err != nil {
		return ""
	}
	return string(payload)
}

func userIDString(value any) (string, bool) {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed), strings.TrimSpace(typed) != ""
	case json.Number:
		if i, err := typed.Int64(); err == nil {
			return fmt.Sprintf("%d", i), true
		}
		return strings.TrimSpace(typed.String()), strings.TrimSpace(typed.String()) != ""
	case float64:
		if typed == float64(int64(typed)) {
			return fmt.Sprintf("%d", int64(typed)), true
		}
		return "", false
	case int:
		return fmt.Sprintf("%d", typed), true
	case int64:
		return fmt.Sprintf("%d", typed), true
	default:
		return "", false
	}
}

func decodeJSONBody(body []byte) map[string]any {
	payload := map[string]any{}
	if len(body) == 0 {
		return payload
	}
	_ = json.Unmarshal(body, &payload)
	return payload
}

func extractModels(body []byte) []domain.UpstreamSyncedModel {
	payload := decodeJSONBody(body)
	rawData, ok := payload["data"].([]any)
	if !ok {
		return nil
	}
	models := make([]domain.UpstreamSyncedModel, 0, len(rawData))
	for _, raw := range rawData {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		id, _ := item["id"].(string)
		if id == "" {
			id, _ = item["name"].(string)
		}
		if id == "" {
			continue
		}
		models = append(models, domain.UpstreamSyncedModel{
			NormalizedModelName:    id,
			UpstreamModelName:      id,
			DisplayName:            displayName(item, id),
			NativeWireProtocol:     protocolForModel(id),
			SupportedWireProtocols: supportedProtocolsForModel(id),
			Capabilities:           capabilitiesForModel(id),
			Status:                 "active",
			RawMetadata:            item,
		})
	}
	return models
}

func applyModelPricingProtocols(models []domain.UpstreamSyncedModel, pricing map[string][]domain.Protocol) []domain.UpstreamSyncedModel {
	if len(pricing) == 0 {
		return models
	}
	for i := range models {
		protocols := pricing[models[i].UpstreamModelName]
		if len(protocols) == 0 {
			protocols = pricing[models[i].NormalizedModelName]
		}
		if len(protocols) == 0 {
			continue
		}
		models[i].SupportedWireProtocols = protocols
		models[i].NativeWireProtocol = protocols[0]
		models[i].Capabilities = capabilitiesForProtocols(models[i].UpstreamModelName, protocols)
	}
	return models
}

func displayName(item map[string]any, fallback string) string {
	for _, key := range []string{"display_name", "name"} {
		if value, ok := item[key].(string); ok && value != "" {
			return value
		}
	}
	return fallback
}

func protocolForModel(model string) domain.Protocol {
	lower := strings.ToLower(model)
	switch {
	case strings.Contains(lower, "claude"):
		return domain.ProtocolAnthropicMessages
	case strings.Contains(lower, "gemini"):
		return domain.ProtocolGeminiGenerate
	default:
		return domain.ProtocolOpenAIChat
	}
}

func supportedProtocolsForModel(model string) []domain.Protocol {
	native := protocolForModel(model)
	if native == domain.ProtocolOpenAIChat {
		return []domain.Protocol{domain.ProtocolOpenAIChat, domain.ProtocolOpenAIResponses}
	}
	return []domain.Protocol{native}
}

func capabilitiesForProtocols(model string, protocols []domain.Protocol) []domain.Capability {
	seen := map[domain.Capability]bool{}
	capabilities := []domain.Capability{}
	add := func(capability domain.Capability) {
		if !seen[capability] {
			seen[capability] = true
			capabilities = append(capabilities, capability)
		}
	}
	for _, protocol := range protocols {
		switch protocol {
		case domain.ProtocolOpenAIResponses:
			add(domain.CapabilityResponses)
		case domain.ProtocolOpenAIChat, domain.ProtocolAnthropicMessages, domain.ProtocolGeminiGenerate:
			add(domain.CapabilityChat)
			add(domain.CapabilityStreaming)
		}
	}
	if strings.Contains(strings.ToLower(model), "vision") || strings.Contains(strings.ToLower(model), "gpt-4o") || strings.Contains(strings.ToLower(model), "gemini") {
		add(domain.CapabilityVision)
	}
	if len(capabilities) == 0 {
		return capabilitiesForModel(model)
	}
	return capabilities
}

func capabilitiesForModel(model string) []domain.Capability {
	capabilities := []domain.Capability{domain.CapabilityChat, domain.CapabilityStreaming}
	lower := strings.ToLower(model)
	if strings.Contains(lower, "embedding") || strings.Contains(lower, "embed") {
		return []domain.Capability{domain.CapabilityEmbedding}
	}
	if strings.Contains(lower, "vision") || strings.Contains(lower, "gpt-4o") || strings.Contains(lower, "gemini") {
		capabilities = append(capabilities, domain.CapabilityVision)
	}
	return capabilities
}

func classifyStatus(statusCode int, body []byte) (domain.UpstreamErrorClass, string) {
	message := messageFromBody(body)
	switch {
	case statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden:
		return domain.UpstreamErrorCredentialExpired, message
	case statusCode == http.StatusTooManyRequests:
		return domain.UpstreamErrorRateLimit, message
	case statusCode >= 500:
		return domain.UpstreamErrorUpstream5xx, message
	case statusCode >= 400 && requiresHumanAction(message):
		return domain.UpstreamErrorActionRequired, message
	case statusCode >= 400:
		return domain.UpstreamErrorInvalidResponse, message
	default:
		return "", ""
	}
}

func messageFromBody(body []byte) string {
	payload := decodeJSONBody(body)
	for _, key := range []string{"message", "error", "detail"} {
		if value, ok := payload[key].(string); ok && value != "" {
			return value
		}
	}
	if errorPayload, ok := payload["error"].(map[string]any); ok {
		if value, ok := errorPayload["message"].(string); ok {
			return value
		}
	}
	bodyText := strings.TrimSpace(string(body))
	if len(bodyText) > 200 {
		return bodyText[:200]
	}
	return bodyText
}

func requiresHumanAction(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "turnstile") ||
		strings.Contains(lower, "captcha") ||
		strings.Contains(lower, "verification") ||
		strings.Contains(lower, "qr")
}

func authMessageIndicatesCredentialProblem(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "access token invalid") ||
		strings.Contains(lower, "access_token_invalid") ||
		strings.Contains(lower, "invalid token") ||
		strings.Contains(lower, "token invalid") ||
		strings.Contains(lower, "token expired") ||
		strings.Contains(lower, "unauthorized") ||
		strings.Contains(lower, "forbidden") ||
		strings.Contains(lower, "new-api-user") ||
		strings.Contains(lower, "user id") ||
		strings.Contains(lower, "mismatch") ||
		strings.Contains(lower, "未登录") ||
		strings.Contains(lower, "无权") ||
		strings.Contains(lower, "不匹配")
}

func responseSucceeded(body []byte) bool {
	payload := decodeJSONBody(body)
	success, ok := payload["success"].(bool)
	if ok {
		return success
	}
	if code, ok := numberFromAny(payload["code"]); ok {
		return code == 0
	}
	return true
}

func responseExplicitlyFailed(body []byte) bool {
	payload := decodeJSONBody(body)
	success, ok := payload["success"].(bool)
	if ok {
		return !success
	}
	if code, ok := numberFromAny(payload["code"]); ok {
		return code != 0
	}
	return false
}

func healthyAPIStatus(accountID string, modelCount int, latencyMS int) domain.UpstreamAccountStatus {
	return domain.UpstreamAccountStatus{
		UpstreamAccountID: accountID,
		APIStatus:         domain.UpstreamAPIStatusHealthy,
		AccountStatus:     domain.AccountCredentialStatusNotConfigured,
		CheckinStatus:     domain.CheckinStatusUnsupported,
		ModelCount:        modelCount,
		LatencyMS:         latencyMS,
		UpdatedAt:         time.Now(),
	}
}

func failedAPIStatus(accountID string, statusCode int, body []byte) domain.UpstreamAccountStatus {
	errorClass, message := classifyStatus(statusCode, body)
	return domain.UpstreamAccountStatus{
		UpstreamAccountID: accountID,
		APIStatus:         domain.UpstreamAPIStatusFailed,
		AccountStatus:     domain.AccountCredentialStatusNotConfigured,
		CheckinStatus:     domain.CheckinStatusUnsupported,
		LastErrorClass:    errorClass,
		LastErrorMessage:  message,
		UpdatedAt:         time.Now(),
	}
}

func failedTransportStatus(accountID string, err error) domain.UpstreamAccountStatus {
	return domain.UpstreamAccountStatus{
		UpstreamAccountID: accountID,
		APIStatus:         domain.UpstreamAPIStatusFailed,
		AccountStatus:     domain.AccountCredentialStatusNotConfigured,
		CheckinStatus:     domain.CheckinStatusUnsupported,
		LastErrorClass:    domain.UpstreamErrorTransport,
		LastErrorMessage:  err.Error(),
		UpdatedAt:         time.Now(),
	}
}

func missingCredentialStatus(accountID string) domain.UpstreamAccountStatus {
	return domain.UpstreamAccountStatus{
		UpstreamAccountID: accountID,
		APIStatus:         domain.UpstreamAPIStatusUnknown,
		AccountStatus:     domain.AccountCredentialStatusNotConfigured,
		CheckinStatus:     domain.CheckinStatusNotConfigured,
		LastErrorClass:    domain.UpstreamErrorCredentialMissing,
		LastErrorMessage:  "account credential is not configured",
		UpdatedAt:         time.Now(),
	}
}

func quotaAmount(body []byte) float64 {
	payload := decodeJSONBody(body)
	for _, key := range []string{"remaining", "remain_quota", "balance", "quota"} {
		if value, ok := numberFromAny(payload[key]); ok {
			return value
		}
	}
	if data, ok := payload["data"].(map[string]any); ok {
		for _, key := range []string{"remaining", "remain_quota", "balance", "quota"} {
			if value, ok := numberFromAny(data[key]); ok {
				return value
			}
		}
	}
	return 0
}

func newAPISelfQuotaAmount(body []byte) float64 {
	payload := decodeJSONBody(body)
	candidates := []map[string]any{payload}
	if data, ok := payload["data"].(map[string]any); ok {
		candidates = append([]map[string]any{data}, candidates...)
	}
	for _, candidate := range candidates {
		quota, hasQuota := numberFromAny(candidate["quota"])
		usedQuota, hasUsedQuota := numberFromAny(candidate["used_quota"])
		if hasQuota && hasUsedQuota {
			return quota - usedQuota
		}
		if hasQuota {
			return quota
		}
	}
	return quotaAmount(body)
}

func numberFromAny(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case int:
		return float64(typed), true
	case json.Number:
		number, err := typed.Float64()
		return number, err == nil
	default:
		return 0, false
	}
}

func latencyMS(start time.Time) int {
	return int(time.Since(start).Milliseconds())
}

func authBearer(token string) string {
	if strings.TrimSpace(token) == "" {
		return ""
	}
	return "Bearer " + strings.TrimSpace(token)
}

func errUnexpectedStatus(statusCode int, body []byte) error {
	errorClass, message := classifyStatus(statusCode, body)
	if message == "" {
		message = fmt.Sprintf("upstream returned status %d", statusCode)
	}
	return fmt.Errorf("%s: %s", errorClass, message)
}
