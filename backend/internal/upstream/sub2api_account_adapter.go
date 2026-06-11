package upstream

import (
	"context"
	"net/http"
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

func (a *Sub2APIAccountAdapter) TestAccountCredential(ctx context.Context, account domain.UpstreamAccount, credential string) (domain.UpstreamAccountStatus, error) {
	if credential == "" {
		return missingCredentialStatus(account.ID), nil
	}
	start := time.Now()
	resp, err := a.do(ctx, http.MethodGet, account, "/api/v1/user/profile", authBearer(credential), nil)
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

func (a *Sub2APIAccountAdapter) RefreshQuota(ctx context.Context, account domain.UpstreamAccount, apiKey string, _ string) (domain.QuotaRefreshResult, error) {
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
