package domain

import "testing"

func TestUpstreamAccountSeparatesAPICredentialFromAccountCredential(t *testing.T) {
	account := UpstreamAccount{
		ID:                    "acct_1",
		Name:                  "API Station A",
		PlatformKind:          PlatformKindNewAPI,
		BaseURL:               "https://api.example.com",
		APIKeyPrefix:          "sk-live",
		AccountCredentialKind: CredentialKindNone,
		Enabled:               true,
		IncludeInRouting:      true,
	}

	if !account.HasAPICredential() {
		t.Fatal("expected API credential to be configured")
	}
	if account.HasAccountCredential() {
		t.Fatal("expected account credential to be absent")
	}
}

func TestUpstreamAccountStatusTreatsMissingAccountCredentialAsNotConfigured(t *testing.T) {
	status := UpstreamAccountStatus{
		APIStatus:     UpstreamAPIStatusHealthy,
		AccountStatus: AccountCredentialStatusNotConfigured,
		CheckinStatus: CheckinStatusNotConfigured,
	}

	if status.NeedsManualAction() {
		t.Fatal("missing optional account credential must not be manual action")
	}
	if !status.CanRouteTraffic() {
		t.Fatal("healthy API status should allow traffic")
	}
}

func TestActionRequiredIsManualActionButNotAPIFailure(t *testing.T) {
	status := UpstreamAccountStatus{
		APIStatus:            UpstreamAPIStatusHealthy,
		AccountStatus:        AccountCredentialStatusValid,
		CheckinStatus:        CheckinStatusActionRequired,
		ActionRequiredReason: "turnstile_required",
	}

	if !status.NeedsManualAction() {
		t.Fatal("action_required should request admin attention")
	}
	if !status.CanRouteTraffic() {
		t.Fatal("check-in challenge must not block API routing")
	}
}
