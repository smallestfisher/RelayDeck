package upstream

import (
	"context"
	"net/http"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/domain"
)

type NewAPIAccountAdapter struct {
	accountHTTPAdapter
}

func NewNewAPIAccountAdapter(httpClient *http.Client, timeout time.Duration) *NewAPIAccountAdapter {
	return &NewAPIAccountAdapter{accountHTTPAdapter: newAccountHTTPAdapter(httpClient, timeout)}
}

func (a *NewAPIAccountAdapter) TestAPI(ctx context.Context, account domain.UpstreamAccount, apiKey string) (domain.UpstreamAccountStatus, error) {
	_, status, err := a.SyncModels(ctx, account, apiKey)
	return status, err
}

func (a *NewAPIAccountAdapter) TestAccountCredential(ctx context.Context, account domain.UpstreamAccount, credential string) (domain.UpstreamAccountStatus, error) {
	if credential == "" {
		return missingCredentialStatus(account.ID), nil
	}
	start := time.Now()
	resp, err := a.doCookie(ctx, http.MethodGet, account, "/api/user/self", credential, nil)
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
		if errorClass == domain.UpstreamErrorActionRequired {
			accountStatus = domain.AccountCredentialStatusActionRequired
		}
		return domain.UpstreamAccountStatus{
			UpstreamAccountID:    account.ID,
			APIStatus:            domain.UpstreamAPIStatusUnknown,
			AccountStatus:        accountStatus,
			CheckinStatus:        domain.CheckinStatusNotConfigured,
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
		CheckinStatus:        domain.CheckinStatusUnchecked,
		LatencyMS:            latencyMS(start),
		LastAccountCheckedAt: time.Now(),
		UpdatedAt:            time.Now(),
	}, nil
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

func (a *NewAPIAccountAdapter) RefreshQuota(ctx context.Context, account domain.UpstreamAccount, apiKey string, _ string) (domain.QuotaRefreshResult, error) {
	start := time.Now()
	resp, err := a.do(ctx, http.MethodGet, account, "/api/usage/token", authBearer(apiKey), nil)
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

func (a *NewAPIAccountAdapter) Checkin(ctx context.Context, account domain.UpstreamAccount, credential string) (domain.CheckinResult, error) {
	if credential == "" {
		return domain.CheckinResult{Status: domain.CheckinStatusNotConfigured, Message: "account credential is not configured"}, nil
	}
	resp, err := a.doCookie(ctx, http.MethodPost, account, "/api/user/checkin", credential, nil)
	if err != nil {
		return domain.CheckinResult{Status: domain.CheckinStatusFailed, Message: err.Error()}, nil
	}
	if resp.statusCode < 200 || resp.statusCode >= 300 {
		errorClass, message := classifyStatus(resp.statusCode, resp.body)
		if errorClass == domain.UpstreamErrorActionRequired {
			return domain.CheckinResult{Status: domain.CheckinStatusActionRequired, Message: message}, nil
		}
		if errorClass == domain.UpstreamErrorCredentialExpired {
			return domain.CheckinResult{Status: domain.CheckinStatusNotConfigured, Message: message}, nil
		}
		return domain.CheckinResult{Status: domain.CheckinStatusFailed, Message: errUnexpectedStatus(resp.statusCode, resp.body).Error()}, nil
	}
	return domain.CheckinResult{Status: domain.CheckinStatusChecked, Message: "checked"}, nil
}
