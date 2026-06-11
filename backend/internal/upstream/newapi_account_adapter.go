package upstream

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/domain"
)

type NewAPIAccountAdapter struct {
	accountHTTPAdapter
}

const defaultNewAPIQuotaPerUnit = 500000.0

func NewNewAPIAccountAdapter(httpClient *http.Client, timeout time.Duration) *NewAPIAccountAdapter {
	return &NewAPIAccountAdapter{accountHTTPAdapter: newAccountHTTPAdapter(httpClient, timeout)}
}

func (a *NewAPIAccountAdapter) TestAPI(ctx context.Context, account domain.UpstreamAccount, apiKey string) (domain.UpstreamAccountStatus, error) {
	_, status, err := a.SyncModels(ctx, account, apiKey)
	return status, err
}

func (a *NewAPIAccountAdapter) TestAccountCredential(ctx context.Context, account domain.UpstreamAccount, credential string) (domain.AccountCredentialTestResult, error) {
	if credential == "" {
		return domain.AccountCredentialTestResult{Status: missingCredentialStatus(account.ID)}, nil
	}
	start := time.Now()
	resp, err := a.doAccountCredential(ctx, http.MethodGet, account, "/api/user/self", credential)
	if err != nil {
		return domain.AccountCredentialTestResult{Status: accountCredentialParseOrTransportStatus(account.ID, err)}, nil
	}
	return domain.AccountCredentialTestResult{Status: newAPIAccountStatusFromResponse(account.ID, resp, latencyMS(start))}, nil
}

func (a *NewAPIAccountAdapter) SyncModels(ctx context.Context, account domain.UpstreamAccount, apiKey string) (domain.ModelSyncResult, domain.UpstreamAccountStatus, error) {
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

func (a *NewAPIAccountAdapter) RefreshQuota(ctx context.Context, account domain.UpstreamAccount, apiKey string, accountCredential string) (domain.QuotaRefreshResult, error) {
	if strings.TrimSpace(accountCredential) != "" && account.AccountCredentialKind == domain.CredentialKindNewAPIAccessToken {
		start := time.Now()
		resp, err := a.doAccountCredential(ctx, http.MethodGet, account, "/api/user/self", accountCredential)
		if err != nil {
			return domain.QuotaRefreshResult{Status: accountCredentialParseOrTransportStatus(account.ID, err)}, nil
		}
		if resp.statusCode < 200 || resp.statusCode >= 300 || responseExplicitlyFailed(resp.body) {
			status := newAPIAccountStatusFromResponse(account.ID, resp, latencyMS(start))
			return domain.QuotaRefreshResult{Status: status}, nil
		}
		amount := newAPIQuotaUnitsToUSD(newAPISelfQuotaAmount(resp.body), a.quotaPerUnit(ctx, account))
		status := healthyAPIStatus(account.ID, 0, latencyMS(start))
		status.AccountStatus = domain.AccountCredentialStatusValid
		status.BalanceAmount = amount
		status.BalanceUnit = "usd"
		status.LastAccountCheckedAt = time.Now()
		return domain.QuotaRefreshResult{Status: status, BalanceAmount: amount, BalanceUnit: "usd"}, nil
	}

	start := time.Now()
	resp, err := a.do(ctx, http.MethodGet, account, "/api/usage/token", authBearer(apiKey), nil)
	if err != nil {
		return domain.QuotaRefreshResult{Status: failedTransportStatus(account.ID, err)}, nil
	}
	if resp.statusCode < 200 || resp.statusCode >= 300 {
		return domain.QuotaRefreshResult{Status: failedAPIStatus(account.ID, resp.statusCode, resp.body)}, nil
	}
	amount := newAPIQuotaUnitsToUSD(quotaAmount(resp.body), a.quotaPerUnit(ctx, account))
	status := healthyAPIStatus(account.ID, 0, latencyMS(start))
	status.BalanceAmount = amount
	status.BalanceUnit = "usd"
	status.LastAPICheckedAt = time.Now()
	return domain.QuotaRefreshResult{Status: status, BalanceAmount: amount, BalanceUnit: "usd"}, nil
}

func (a *NewAPIAccountAdapter) Checkin(ctx context.Context, account domain.UpstreamAccount, credential string) (domain.CheckinResult, error) {
	if credential == "" {
		return domain.CheckinResult{Status: domain.CheckinStatusNotConfigured, Message: "account credential is not configured"}, nil
	}
	resp, err := a.doAccountCredential(ctx, http.MethodPost, account, "/api/user/checkin", credential)
	if err != nil {
		return checkinResultForCredentialError(err), nil
	}
	if resp.statusCode < 200 || resp.statusCode >= 300 || responseExplicitlyFailed(resp.body) {
		errorClass, message := classifyStatus(resp.statusCode, resp.body)
		if resp.statusCode >= 200 && resp.statusCode < 300 {
			message = messageFromBody(resp.body)
			switch {
			case authMessageIndicatesCredentialProblem(message):
				errorClass = domain.UpstreamErrorCredentialExpired
			case requiresHumanAction(message):
				errorClass = domain.UpstreamErrorActionRequired
			default:
				errorClass = domain.UpstreamErrorInvalidResponse
			}
		}
		if errorClass == domain.UpstreamErrorActionRequired {
			return domain.CheckinResult{
				Status:               domain.CheckinStatusActionRequired,
				Message:              message,
				LastErrorClass:       errorClass,
				LastErrorMessage:     message,
				ActionRequiredReason: message,
			}, nil
		}
		if errorClass == domain.UpstreamErrorCredentialExpired {
			return domain.CheckinResult{
				Status:               domain.CheckinStatusActionRequired,
				Message:              message,
				AccountStatus:        domain.AccountCredentialStatusExpired,
				LastErrorClass:       errorClass,
				LastErrorMessage:     message,
				ActionRequiredReason: message,
			}, nil
		}
		return domain.CheckinResult{Status: domain.CheckinStatusFailed, Message: errUnexpectedStatus(resp.statusCode, resp.body).Error()}, nil
	}
	return domain.CheckinResult{Status: domain.CheckinStatusChecked, Message: "checked"}, nil
}

func (a *NewAPIAccountAdapter) doAccountCredential(ctx context.Context, method string, account domain.UpstreamAccount, path string, credential string) (adapterResponse, error) {
	if account.AccountCredentialKind == domain.CredentialKindNewAPIAccessToken {
		parsed, err := parseNewAPIAccessCredential(credential)
		if err != nil {
			return adapterResponse{}, err
		}
		return a.doNewAPIUser(ctx, method, account, path, parsed, nil)
	}
	return a.doCookie(ctx, method, account, path, credential, nil)
}

func (a *NewAPIAccountAdapter) quotaPerUnit(ctx context.Context, account domain.UpstreamAccount) float64 {
	resp, err := a.do(ctx, http.MethodGet, account, "/api/status", "", nil)
	if err != nil || resp.statusCode < 200 || resp.statusCode >= 300 {
		return defaultNewAPIQuotaPerUnit
	}
	if value, ok := numberFromAny(decodeJSONBody(resp.body)["quota_per_unit"]); ok && value > 0 {
		return value
	}
	return defaultNewAPIQuotaPerUnit
}

func newAPIQuotaUnitsToUSD(quotaUnits float64, quotaPerUnit float64) float64 {
	if quotaPerUnit <= 0 {
		quotaPerUnit = defaultNewAPIQuotaPerUnit
	}
	return quotaUnits / quotaPerUnit
}

func newAPIAccountStatusFromResponse(accountID string, resp adapterResponse, latency int) domain.UpstreamAccountStatus {
	status := domain.UpstreamAccountStatus{
		UpstreamAccountID:    accountID,
		APIStatus:            domain.UpstreamAPIStatusUnknown,
		AccountStatus:        domain.AccountCredentialStatusValid,
		CheckinStatus:        domain.CheckinStatusUnchecked,
		LatencyMS:            latency,
		LastAccountCheckedAt: time.Now(),
		UpdatedAt:            time.Now(),
	}
	if resp.statusCode >= 200 && resp.statusCode < 300 && responseSucceeded(resp.body) {
		return status
	}
	errorClass, message := classifyStatus(resp.statusCode, resp.body)
	if resp.statusCode >= 200 && resp.statusCode < 300 && responseExplicitlyFailed(resp.body) {
		message = messageFromBody(resp.body)
		switch {
		case authMessageIndicatesCredentialProblem(message):
			errorClass = domain.UpstreamErrorCredentialExpired
		case requiresHumanAction(message):
			errorClass = domain.UpstreamErrorActionRequired
		default:
			errorClass = domain.UpstreamErrorInvalidResponse
		}
	}
	status.AccountStatus = domain.AccountCredentialStatusFailed
	status.CheckinStatus = domain.CheckinStatusNotConfigured
	status.LastErrorClass = errorClass
	status.LastErrorMessage = message
	if errorClass == domain.UpstreamErrorCredentialExpired {
		status.AccountStatus = domain.AccountCredentialStatusExpired
	}
	if errorClass == domain.UpstreamErrorActionRequired {
		status.AccountStatus = domain.AccountCredentialStatusActionRequired
		status.ActionRequiredReason = message
	}
	return status
}

func accountCredentialParseOrTransportStatus(accountID string, err error) domain.UpstreamAccountStatus {
	status := failedTransportStatus(accountID, err)
	status.APIStatus = domain.UpstreamAPIStatusUnknown
	status.AccountStatus = domain.AccountCredentialStatusFailed
	status.CheckinStatus = domain.CheckinStatusNotConfigured
	status.LastErrorClass = domain.UpstreamErrorInvalidResponse
	if strings.Contains(err.Error(), "not configured") {
		status = missingCredentialStatus(accountID)
	}
	return status
}

func checkinResultForCredentialError(err error) domain.CheckinResult {
	if strings.Contains(err.Error(), "not configured") {
		return domain.CheckinResult{Status: domain.CheckinStatusNotConfigured, Message: err.Error(), LastErrorClass: domain.UpstreamErrorCredentialMissing, LastErrorMessage: err.Error()}
	}
	return domain.CheckinResult{
		Status:           domain.CheckinStatusFailed,
		Message:          err.Error(),
		AccountStatus:    domain.AccountCredentialStatusFailed,
		LastErrorClass:   domain.UpstreamErrorInvalidResponse,
		LastErrorMessage: err.Error(),
	}
}

func (a *NewAPIAccountAdapter) TestCall(ctx context.Context, account domain.UpstreamAccount, apiKey string, modelName string, protocol string, streaming bool, message string) (map[string]any, error) {
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
			"prompt": message,
			"stream": streaming,
		}
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	resp, err := a.do(ctx, http.MethodPost, account, path, authBearer(apiKey), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	if resp.statusCode < 200 || resp.statusCode >= 300 {
		return map[string]any{"error": string(resp.body), "status_code": resp.statusCode}, nil
	}

	var result map[string]any
	if err := json.Unmarshal(resp.body, &result); err != nil {
		return map[string]any{"raw": string(resp.body)}, nil
	}
	return result, nil
}
