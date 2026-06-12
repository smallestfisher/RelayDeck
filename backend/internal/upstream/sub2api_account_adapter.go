package upstream

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/domain"
)

type Sub2APIAccountAdapter struct {
	accountHTTPAdapter
}

func NewSub2APIAccountAdapter(httpClient *http.Client, timeout time.Duration) *Sub2APIAccountAdapter {
	return &Sub2APIAccountAdapter{accountHTTPAdapter: newAccountHTTPAdapter(httpClient, timeout)}
}

func (a *Sub2APIAccountAdapter) TestAPI(ctx context.Context, account domain.UpstreamAccount, apiKey string) (domain.UpstreamAccountStatus, error) {
	_, status, err := a.SyncModels(ctx, account, apiKey)
	return status, err
}

func (a *Sub2APIAccountAdapter) TestAccountCredential(ctx context.Context, account domain.UpstreamAccount, credential string) (domain.AccountCredentialTestResult, error) {
	if credential == "" {
		return domain.AccountCredentialTestResult{Status: missingCredentialStatus(account.ID)}, nil
	}
	if shouldUseSub2APIRefreshCredential(account, credential) {
		refresh, status, err := a.refreshAccessToken(ctx, account, credential)
		if err != nil || status.LastErrorClass != "" {
			return domain.AccountCredentialTestResult{Status: status, CredentialUpdate: refresh.CredentialUpdate}, nil
		}
		status, err = a.testProfileWithAccessToken(ctx, account, refresh.AccessToken)
		return domain.AccountCredentialTestResult{Status: status, CredentialUpdate: refresh.CredentialUpdate}, err
	}
	status, err := a.testProfileWithAccessToken(ctx, account, credential)
	return domain.AccountCredentialTestResult{Status: status}, err
}

func (a *Sub2APIAccountAdapter) testProfileWithAccessToken(ctx context.Context, account domain.UpstreamAccount, accessToken string) (domain.UpstreamAccountStatus, error) {
	start := time.Now()
	resp, err := a.do(ctx, http.MethodGet, account, "/api/v1/user/profile", authBearer(accessToken), nil)
	if err != nil {
		status := failedTransportStatus(account.ID, err)
		status.AccountStatus = domain.AccountCredentialStatusFailed
		return status, nil
	}
	if resp.statusCode < 200 || resp.statusCode >= 300 {
		errorClass, message := classifyStatus(resp.statusCode, resp.body)
		accountStatus := domain.AccountCredentialStatusFailed
		if errorClass == domain.UpstreamErrorCredentialExpired {
			accountStatus = domain.AccountCredentialStatusExpired
		}
		return domain.UpstreamAccountStatus{
			UpstreamAccountID:    account.ID,
			APIStatus:            domain.UpstreamAPIStatusUnknown,
			AccountStatus:        accountStatus,
			CheckinStatus:        domain.CheckinStatusUnsupported,
			LatencyMS:            latencyMS(start),
			LastAccountCheckedAt: time.Now(),
			LastErrorClass:       errorClass,
			LastErrorMessage:     message,
			UpdatedAt:            time.Now(),
		}, nil
	}
	return domain.UpstreamAccountStatus{
		UpstreamAccountID:    account.ID,
		APIStatus:            domain.UpstreamAPIStatusUnknown,
		AccountStatus:        domain.AccountCredentialStatusValid,
		CheckinStatus:        domain.CheckinStatusUnsupported,
		LatencyMS:            latencyMS(start),
		LastAccountCheckedAt: time.Now(),
		UpdatedAt:            time.Now(),
	}, nil
}

func (a *Sub2APIAccountAdapter) SyncModels(ctx context.Context, account domain.UpstreamAccount, apiKey string) (domain.ModelSyncResult, domain.UpstreamAccountStatus, error) {
	start := time.Now()
	resp, err := a.do(ctx, http.MethodGet, account, "/v1/models", authBearer(apiKey), nil)
	if err != nil {
		status := failedTransportStatus(account.ID, err)
		return domain.ModelSyncResult{AccountID: account.ID}, status, nil
	}
	if resp.statusCode < 200 || resp.statusCode >= 300 {
		status := failedAPIStatus(account.ID, resp.statusCode, resp.body)
		return domain.ModelSyncResult{AccountID: account.ID}, status, nil
	}
	models := extractModels(resp.body)
	status := healthyAPIStatus(account.ID, len(models), latencyMS(start))
	status.LastAPICheckedAt = time.Now()
	status.LastModelSyncedAt = time.Now()
	return domain.ModelSyncResult{
		AccountID:     account.ID,
		CreatedModels: len(models),
		UpdatedModels: len(models),
		Models:        models,
	}, status, nil
}

func (a *Sub2APIAccountAdapter) RefreshQuota(ctx context.Context, account domain.UpstreamAccount, apiKey string, accountCredential string) (domain.QuotaRefreshResult, error) {
	if strings.TrimSpace(accountCredential) != "" && shouldUseSub2APIRefreshCredential(account, accountCredential) {
		refresh, status, err := a.refreshAccessToken(ctx, account, accountCredential)
		if err != nil || status.LastErrorClass != "" {
			return domain.QuotaRefreshResult{Status: status, CredentialUpdate: refresh.CredentialUpdate}, nil
		}
		start := time.Now()
		resp, err := a.do(ctx, http.MethodGet, account, "/api/v1/user/platform-quotas", authBearer(refresh.AccessToken), nil)
		if err != nil {
			status := failedTransportStatus(account.ID, err)
			status.AccountStatus = domain.AccountCredentialStatusFailed
			return domain.QuotaRefreshResult{Status: status, CredentialUpdate: refresh.CredentialUpdate}, nil
		}
		if resp.statusCode < 200 || resp.statusCode >= 300 {
			status := sub2APIAccountStatusFromFailure(account.ID, resp.statusCode, resp.body, latencyMS(start))
			return domain.QuotaRefreshResult{Status: status, CredentialUpdate: refresh.CredentialUpdate}, nil
		}
		amount := sub2APIPlatformQuotaAmount(resp.body)
		status = healthyAPIStatus(account.ID, 0, latencyMS(start))
		status.AccountStatus = domain.AccountCredentialStatusValid
		status.BalanceAmount = amount
		status.BalanceUnit = "usd"
		status.LastAccountCheckedAt = time.Now()
		return domain.QuotaRefreshResult{Status: status, BalanceAmount: amount, BalanceUnit: "usd", CredentialUpdate: refresh.CredentialUpdate}, nil
	}

	start := time.Now()
	resp, err := a.do(ctx, http.MethodGet, account, "/v1/usage", authBearer(apiKey), nil)
	if err != nil {
		return domain.QuotaRefreshResult{Status: failedTransportStatus(account.ID, err)}, nil
	}
	if resp.statusCode < 200 || resp.statusCode >= 300 {
		return domain.QuotaRefreshResult{Status: failedAPIStatus(account.ID, resp.statusCode, resp.body)}, nil
	}
	amount := quotaAmount(resp.body)
	status := healthyAPIStatus(account.ID, 0, latencyMS(start))
	status.BalanceAmount = amount
	status.BalanceUnit = "quota"
	status.LastAPICheckedAt = time.Now()
	return domain.QuotaRefreshResult{Status: status, BalanceAmount: amount, BalanceUnit: "quota"}, nil
}

func (a *Sub2APIAccountAdapter) Checkin(context.Context, domain.UpstreamAccount, string) (domain.CheckinResult, error) {
	return domain.CheckinResult{Status: domain.CheckinStatusUnsupported, Message: "sub2api check-in is not supported"}, nil
}

type sub2APIRefreshResult struct {
	AccessToken      string
	RefreshToken     string
	CredentialUpdate *domain.UpstreamCredentialUpdate
}

func (a *Sub2APIAccountAdapter) refreshAccessToken(ctx context.Context, account domain.UpstreamAccount, credential string) (sub2APIRefreshResult, domain.UpstreamAccountStatus, error) {
	start := time.Now()
	parsed, err := parseSub2APIRefreshCredential(credential)
	if err != nil {
		status := accountCredentialParseOrTransportStatus(account.ID, err)
		return sub2APIRefreshResult{}, status, nil
	}
	payload, err := json.Marshal(map[string]string{"refresh_token": parsed.RefreshToken})
	if err != nil {
		status := accountCredentialParseOrTransportStatus(account.ID, err)
		return sub2APIRefreshResult{}, status, nil
	}
	resp, err := a.do(ctx, http.MethodPost, account, "/api/v1/auth/refresh", "", bytes.NewReader(payload))
	if err != nil {
		status := failedTransportStatus(account.ID, err)
		status.APIStatus = domain.UpstreamAPIStatusUnknown
		status.AccountStatus = domain.AccountCredentialStatusFailed
		status.CheckinStatus = domain.CheckinStatusUnsupported
		return sub2APIRefreshResult{}, status, nil
	}
	if resp.statusCode < 200 || resp.statusCode >= 300 || responseExplicitlyFailed(resp.body) {
		status := sub2APIAccountStatusFromFailure(account.ID, resp.statusCode, resp.body, latencyMS(start))
		return sub2APIRefreshResult{}, status, nil
	}
	accessToken := firstStringField(resp.body, "access_token", "token")
	if accessToken == "" {
		status := sub2APIAccountStatus(account.ID, domain.AccountCredentialStatusFailed, domain.UpstreamErrorInvalidResponse, "sub2api refresh response did not include access_token", latencyMS(start))
		return sub2APIRefreshResult{}, status, nil
	}
	refreshToken := firstStringField(resp.body, "refresh_token")
	result := sub2APIRefreshResult{AccessToken: accessToken, RefreshToken: parsed.RefreshToken}
	if refreshToken != "" && refreshToken != parsed.RefreshToken {
		result.RefreshToken = refreshToken
		result.CredentialUpdate = &domain.UpstreamCredentialUpdate{
			Kind:      domain.CredentialKindSub2APIRefreshToken,
			Plaintext: encodeSub2APIRefreshCredential(refreshToken),
		}
	}
	return result, domain.UpstreamAccountStatus{}, nil
}

func shouldUseSub2APIRefreshCredential(account domain.UpstreamAccount, credential string) bool {
	credential = strings.TrimSpace(credential)
	return account.AccountCredentialKind == domain.CredentialKindSub2APIRefreshToken ||
		strings.HasPrefix(credential, "rt_") ||
		strings.Contains(credential, `"refresh_token"`)
}

func sub2APIAccountStatusFromFailure(accountID string, statusCode int, body []byte, latency int) domain.UpstreamAccountStatus {
	errorClass, message := classifyStatus(statusCode, body)
	if statusCode >= 200 && statusCode < 300 && responseExplicitlyFailed(body) {
		message = messageFromBody(body)
		errorClass = domain.UpstreamErrorInvalidResponse
		if authMessageIndicatesCredentialProblem(message) || sub2APIRefreshMessageExpired(message) {
			errorClass = domain.UpstreamErrorCredentialExpired
		}
	}
	accountStatus := domain.AccountCredentialStatusFailed
	if errorClass == domain.UpstreamErrorCredentialExpired {
		accountStatus = domain.AccountCredentialStatusExpired
	}
	return sub2APIAccountStatus(accountID, accountStatus, errorClass, message, latency)
}

func sub2APIAccountStatus(accountID string, accountStatus domain.AccountCredentialStatus, errorClass domain.UpstreamErrorClass, message string, latency int) domain.UpstreamAccountStatus {
	return domain.UpstreamAccountStatus{
		UpstreamAccountID:    accountID,
		APIStatus:            domain.UpstreamAPIStatusUnknown,
		AccountStatus:        accountStatus,
		CheckinStatus:        domain.CheckinStatusUnsupported,
		LatencyMS:            latency,
		LastAccountCheckedAt: time.Now(),
		LastErrorClass:       errorClass,
		LastErrorMessage:     message,
		UpdatedAt:            time.Now(),
	}
}

func sub2APIRefreshMessageExpired(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "refresh") &&
		(strings.Contains(lower, "expired") ||
			strings.Contains(lower, "revoked") ||
			strings.Contains(lower, "reused") ||
			strings.Contains(lower, "invalid"))
}

func firstStringField(body []byte, keys ...string) string {
	payload := decodeJSONBody(body)
	for _, candidate := range []map[string]any{payload, nestedMap(payload, "data")} {
		if candidate == nil {
			continue
		}
		for _, key := range keys {
			if value, ok := candidate[key].(string); ok && strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}
	}
	return ""
}

func nestedMap(payload map[string]any, key string) map[string]any {
	if nested, ok := payload[key].(map[string]any); ok {
		return nested
	}
	return nil
}

func sub2APIPlatformQuotaAmount(body []byte) float64 {
	payload := decodeJSONBody(body)
	for _, candidate := range []map[string]any{nestedMap(payload, "data"), payload} {
		if candidate == nil {
			continue
		}
		rawQuotas, ok := candidate["platform_quotas"].([]any)
		if !ok {
			continue
		}
		total := 0.0
		for _, rawQuota := range rawQuotas {
			quota, ok := rawQuota.(map[string]any)
			if !ok {
				continue
			}
			limit, hasLimit := numberFromAny(quota["monthly_limit_usd"])
			usage, hasUsage := numberFromAny(quota["monthly_usage_usd"])
			if hasLimit {
				if !hasUsage {
					usage = 0
				}
				total += limit - usage
			}
		}
		return total
	}
	if amount := quotaAmount(body); amount != 0 {
		return amount
	}
	return 0
}

func (a *Sub2APIAccountAdapter) TestCall(ctx context.Context, account domain.UpstreamAccount, apiKey string, modelName string, protocol string, streaming bool, message string) (domain.UpstreamTestCallResult, error) {
	start := time.Now()
	var path string
	var reqBody map[string]any

	switch protocol {
	case "openai-chat":
		path = "/v1/chat/completions"
		reqBody = map[string]any{
			"model":    modelName,
			"messages": []map[string]string{{"role": "user", "content": message}},
			"stream":   streaming,
		}
	case "claude-messages":
		path = "/v1/messages"
		reqBody = map[string]any{
			"model":      modelName,
			"messages":   []map[string]string{{"role": "user", "content": message}},
			"max_tokens": 1024,
			"stream":     streaming,
		}
	case "openai-responses":
		path = "/v1/responses"
		reqBody = map[string]any{
			"model":  modelName,
			"input":  message,
			"stream": streaming,
		}
	default:
		return domain.UpstreamTestCallResult{}, errors.New("unsupported protocol: " + protocol)
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return domain.UpstreamTestCallResult{}, err
	}
	resp, err := a.do(ctx, http.MethodPost, account, path, authBearer(apiKey), bytes.NewReader(bodyBytes))
	if err != nil {
		return domain.UpstreamTestCallResult{}, err
	}
	testResult := domain.UpstreamTestCallResult{HTTPStatus: resp.statusCode, Protocol: protocol, OK: resp.statusCode >= 200 && resp.statusCode < 300, LatencyMS: latencyMS(start)}
	if resp.statusCode < 200 || resp.statusCode >= 300 {
		errorClass, message := classifyStatus(resp.statusCode, resp.body)
		testResult.ErrorClass = errorClass
		testResult.ErrorMessage = message
		return testResult, nil
	}
	var upstreamResponse map[string]any
	if err := json.Unmarshal(resp.body, &upstreamResponse); err == nil {
		testResult.UpstreamResponse = upstreamResponse
	}
	return testResult, nil
}
