import type {
  AccountCredentialStatus,
  UpstreamAccount,
  UpstreamAccountEvent,
  UpstreamAccountInput,
  UpstreamAccountPage,
  UpstreamAccountStatusSnapshot,
  UpstreamActionName,
  UpstreamActionResult,
  UpstreamBatchActionName,
  UpstreamCheckinStatus,
  UpstreamCredentialKind,
  UpstreamModel,
  UpstreamPlatformKind,
  UpstreamTestCallResult,
} from '../types';

async function requestJSON<T>(path: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(path, {
    ...init,
    credentials: 'include',
    headers: {
      ...(init.body ? { 'Content-Type': 'application/json' } : {}),
      ...init.headers,
    },
  });
  if (!response.ok) {
    let message = response.statusText;
    try {
      const payload = (await response.json()) as { error?: string };
      message = payload.error ?? message;
    } catch {
      // Keep the HTTP status text when the backend did not return JSON.
    }
    throw new Error(message);
  }
  if (response.status === 204) {
    return undefined as T;
  }
  return (await response.json()) as T;
}

export const adminApi = {
  async listUpstreamAccounts(params: {
    limit?: number;
    offset?: number;
    q?: string;
    platformKind?: string;
    apiStatus?: string;
    accountStatus?: string;
    latencyBand?: string;
  } = {}): Promise<UpstreamAccountPage> {
    const query = new URLSearchParams();
    if (params.limit) query.set('limit', String(params.limit));
    if (params.offset) query.set('offset', String(params.offset));
    if (params.q) query.set('q', params.q);
    if (params.platformKind && params.platformKind !== 'all') query.set('platform_kind', params.platformKind);
    if (params.apiStatus && params.apiStatus !== 'all') query.set('api_status', params.apiStatus);
    if (params.accountStatus && params.accountStatus !== 'all') query.set('account_status', params.accountStatus);
    if (params.latencyBand && params.latencyBand !== 'all') query.set('latency_band', params.latencyBand);
    const path = `/api/admin/upstreams/accounts${query.size > 0 ? `?${query.toString()}` : ''}`;
    const payload = await requestJSON<{
      items: RawUpstreamAccount[];
      total?: number;
      limit?: number;
      offset?: number;
      metrics?: { total: number; healthy: number; warning: number; manual: number };
    }>(path);
    return {
      items: payload.items.map(mapUpstreamAccount),
      total: payload.total ?? payload.items.length,
      limit: payload.limit ?? params.limit ?? payload.items.length,
      offset: payload.offset ?? params.offset ?? 0,
      metrics: payload.metrics,
    };
  },

  async createUpstreamAccount(input: UpstreamAccountInput): Promise<UpstreamAccount> {
    const payload = await requestJSON<RawUpstreamAccount>('/api/admin/upstreams/accounts', {
      method: 'POST',
      body: JSON.stringify(toRawUpstreamAccountInput(input)),
    });
    return mapUpstreamAccount(payload);
  },

  async updateUpstreamAccount(id: string, input: UpstreamAccountInput): Promise<UpstreamAccount> {
    const payload = await requestJSON<RawUpstreamAccount>(`/api/admin/upstreams/accounts/${id}`, {
      method: 'PUT',
      body: JSON.stringify(toRawUpstreamAccountInput(input)),
    });
    return mapUpstreamAccount(payload);
  },

  async deleteUpstreamAccount(id: string): Promise<void> {
    await requestJSON<void>(`/api/admin/upstreams/accounts/${id}`, { method: 'DELETE' });
  },

  async runUpstreamAction(id: string, action: UpstreamActionName): Promise<UpstreamActionResult> {
    const payload = await requestJSON<RawUpstreamActionResult>(`/api/admin/upstreams/accounts/${id}/${action}`, { method: 'POST' });
    return mapUpstreamActionResult(payload);
  },

  async testUpstreamDraft(input: UpstreamAccountInput): Promise<UpstreamActionResult> {
    return requestJSON<UpstreamActionResult>('/api/admin/upstreams/test', {
      method: 'POST',
      body: JSON.stringify({
        platform_kind: input.platformKind,
        base_url: input.baseUrl,
        api_key: input.apiKey,
      }),
    });
  },

  async runBatchUpstreamAction(ids: string[], action: UpstreamBatchActionName): Promise<UpstreamActionResult[]> {
    const payload = await requestJSON<{ results: RawUpstreamActionResult[] }>(`/api/admin/upstreams/accounts/batch/${action}`, {
      method: 'POST',
      body: JSON.stringify({ ids }),
    });
    return payload.results.map(mapUpstreamActionResult);
  },

  async batchDeleteUpstreamAccounts(ids: string[]): Promise<UpstreamActionResult[]> {
    const payload = await requestJSON<{ results: RawUpstreamActionResult[] }>('/api/admin/upstreams/accounts/batch/delete', {
      method: 'POST',
      body: JSON.stringify({ ids }),
    });
    return payload.results.map(mapUpstreamActionResult);
  },

  async listUpstreamModels(id: string): Promise<UpstreamModel[]> {
    const payload = await requestJSON<{ items: RawUpstreamModel[] }>(`/api/admin/upstreams/accounts/${id}/models`);
    return payload.items.map(mapUpstreamModel);
  },

  async listUpstreamEvents(id: string): Promise<UpstreamAccountEvent[]> {
    const payload = await requestJSON<{ items: RawUpstreamAccountEvent[] }>(`/api/admin/upstreams/accounts/${id}/events`);
    return payload.items.map(mapUpstreamEvent);
  },

  async testUpstreamCall(id: string, input: { modelName: string; protocol: string; streaming: boolean; message: string }): Promise<UpstreamTestCallResult> {
    const payload = await requestJSON<RawUpstreamTestCallResult>(`/api/admin/upstreams/accounts/${id}/test-call`, {
      method: 'POST',
      body: JSON.stringify({
        model_name: input.modelName,
        protocol: input.protocol,
        streaming: input.streaming,
        message: input.message,
      }),
    });
    return mapUpstreamTestCallResult(payload);
  },
};

interface RawUpstreamAccount {
  id: string;
  name: string;
  code: string;
  platform_kind: UpstreamPlatformKind;
  base_url: string;
  enabled: boolean;
  include_in_routing: boolean;
  priority: number;
  api_key_prefix: string;
  has_api_credential: boolean;
  account_credential_kind: UpstreamCredentialKind;
  has_account_credential: boolean;
  auto_sync_models: boolean;
  auto_refresh_quota: boolean;
  auto_checkin: boolean;
  note: string;
  status: RawUpstreamStatus;
  created_at: string;
  updated_at: string;
}

interface RawUpstreamStatus {
  UpstreamAccountID?: string;
  upstream_account_id?: string;
  APIStatus?: string;
  api_status?: string;
  AccountStatus?: string;
  account_status?: string;
  CheckinStatus?: string;
  checkin_status?: string;
  ModelCount?: number;
  model_count?: number;
  LatencyMS?: number;
  latency_ms?: number;
  APILatencyMS?: number;
  api_latency_ms?: number;
  BalanceAmount?: number;
  balance_amount?: number;
  BalanceUnit?: string;
  balance_unit?: string;
  LastAPICheckedAt?: string;
  last_api_checked_at?: string;
  LastAccountCheckedAt?: string;
  last_account_checked_at?: string;
  LastModelSyncedAt?: string;
  last_model_synced_at?: string;
  LastCheckinAt?: string;
  last_checkin_at?: string;
  LastErrorClass?: string;
  last_error_class?: string;
  LastErrorMessage?: string;
  last_error_message?: string;
  ActionRequiredReason?: string;
  action_required_reason?: string;
  UpdatedAt?: string;
  updated_at?: string;
}

interface RawUpstreamModel {
  ID?: string;
  id?: string;
  UpstreamAccountID?: string;
  upstream_account_id?: string;
  NormalizedModelName?: string;
  normalized_model_name?: string;
  UpstreamModelName?: string;
  upstream_model_name?: string;
  DisplayName?: string;
  display_name?: string;
  NativeWireProtocol?: string;
  native_wire_protocol?: string;
  SupportedWireProtocols?: string[];
  supported_wire_protocols?: string[];
  Capabilities?: string[];
  capabilities?: string[];
  Status?: string;
  status?: string;
  RawMetadata?: Record<string, unknown>;
  raw_metadata?: Record<string, unknown>;
  LastSyncedAt?: string;
  last_synced_at?: string;
}

interface RawUpstreamAccountEvent {
  ID?: string;
  id?: string;
  UpstreamAccountID?: string;
  upstream_account_id?: string;
  Operation?: string;
  operation?: string;
  Status?: string;
  status?: string;
  ErrorClass?: string;
  error_class?: string;
  Message?: string;
  message?: string;
  LatencyMS?: number;
  latency_ms?: number;
  Metadata?: Record<string, unknown>;
  metadata?: Record<string, unknown>;
  CreatedAt?: string;
  created_at?: string;
}

interface RawUpstreamActionResult {
  id: string;
  status: 'success' | 'failed' | 'not_found';
  message?: string;
  account_status?: RawUpstreamStatus;
}

interface RawUpstreamTestCallResult {
  id: string;
  http_status?: number;
  protocol?: string;
  ok?: boolean;
  message?: string;
  error_class?: string;
  latency_ms?: number;
  account_status?: RawUpstreamStatus;
}

function toRawUpstreamAccountInput(input: UpstreamAccountInput) {
  return {
    name: input.name,
    code: input.code,
    platform_kind: input.platformKind,
    base_url: input.baseUrl,
    enabled: input.enabled,
    include_in_routing: input.includeInRouting,
    priority: input.priority,
    api_key: input.apiKey,
    account_credential_kind: input.accountCredentialKind,
    account_credential: input.accountCredential,
    auto_sync_models: input.autoSyncModels,
    auto_refresh_quota: input.autoRefreshQuota,
    auto_checkin: input.autoCheckin,
    note: input.note,
  };
}

function mapUpstreamActionResult(raw: RawUpstreamActionResult): UpstreamActionResult {
  return {
    id: raw.id,
    status: raw.status,
    message: raw.message,
    accountStatus: raw.account_status ? mapUpstreamStatus(raw.account_status) : undefined,
  };
}

function mapUpstreamTestCallResult(raw: RawUpstreamTestCallResult): UpstreamTestCallResult {
  return {
    id: raw.id,
    httpStatus: raw.http_status ?? 0,
    protocol: raw.protocol ?? '',
    ok: Boolean(raw.ok),
    message: raw.message,
    errorClass: raw.error_class,
    latencyMs: raw.latency_ms ?? 0,
    accountStatus: raw.account_status ? mapUpstreamStatus(raw.account_status) : undefined,
  };
}

function mapUpstreamAccount(raw: RawUpstreamAccount): UpstreamAccount {
  return {
    id: raw.id,
    name: raw.name,
    code: raw.code,
    platformKind: raw.platform_kind,
    baseUrl: raw.base_url,
    enabled: raw.enabled,
    includeInRouting: raw.include_in_routing,
    priority: raw.priority,
    apiKeyPrefix: raw.api_key_prefix,
    hasApiCredential: raw.has_api_credential,
    accountCredentialKind: raw.account_credential_kind,
    hasAccountCredential: raw.has_account_credential,
    autoSyncModels: raw.auto_sync_models,
    autoRefreshQuota: raw.auto_refresh_quota,
    autoCheckin: raw.auto_checkin,
    note: raw.note,
    status: mapUpstreamStatus(raw.status),
    createdAt: raw.created_at,
    updatedAt: raw.updated_at,
  };
}

function mapUpstreamStatus(raw: RawUpstreamStatus): UpstreamAccountStatusSnapshot {
  return {
    upstreamAccountId: raw.upstream_account_id ?? raw.UpstreamAccountID ?? '',
    apiStatus: (raw.api_status ?? raw.APIStatus ?? 'unknown') as UpstreamAccountStatusSnapshot['apiStatus'],
    accountStatus: (raw.account_status ?? raw.AccountStatus ?? 'not_configured') as AccountCredentialStatus,
    checkinStatus: (raw.checkin_status ?? raw.CheckinStatus ?? 'unsupported') as UpstreamCheckinStatus,
    modelCount: raw.model_count ?? raw.ModelCount ?? 0,
    latencyMs: raw.latency_ms ?? raw.LatencyMS ?? 0,
    apiLatencyMs: raw.api_latency_ms ?? raw.APILatencyMS ?? 0,
    balanceAmount: raw.balance_amount ?? raw.BalanceAmount ?? 0,
    balanceUnit: raw.balance_unit ?? raw.BalanceUnit ?? '',
    lastApiCheckedAt: normalizeTimestamp(raw.last_api_checked_at ?? raw.LastAPICheckedAt),
    lastAccountCheckedAt: normalizeTimestamp(raw.last_account_checked_at ?? raw.LastAccountCheckedAt),
    lastModelSyncedAt: normalizeTimestamp(raw.last_model_synced_at ?? raw.LastModelSyncedAt),
    lastCheckinAt: normalizeTimestamp(raw.last_checkin_at ?? raw.LastCheckinAt),
    lastErrorClass: raw.last_error_class ?? raw.LastErrorClass,
    lastErrorMessage: raw.last_error_message ?? raw.LastErrorMessage,
    actionRequiredReason: raw.action_required_reason ?? raw.ActionRequiredReason,
    updatedAt: raw.updated_at ?? raw.UpdatedAt,
  };
}

function normalizeTimestamp(value?: string): string | undefined {
  if (!value || value.startsWith('0001-01-01')) {
    return undefined;
  }
  return value;
}

function mapUpstreamModel(raw: RawUpstreamModel): UpstreamModel {
  return {
    id: raw.id ?? raw.ID ?? '',
    upstreamAccountId: raw.upstream_account_id ?? raw.UpstreamAccountID ?? '',
    normalizedModelName: raw.normalized_model_name ?? raw.NormalizedModelName ?? '',
    upstreamModelName: raw.upstream_model_name ?? raw.UpstreamModelName ?? '',
    displayName: raw.display_name ?? raw.DisplayName ?? '',
    nativeWireProtocol: raw.native_wire_protocol ?? raw.NativeWireProtocol ?? '',
    supportedWireProtocols: raw.supported_wire_protocols ?? raw.SupportedWireProtocols ?? [],
    capabilities: (raw.capabilities ?? raw.Capabilities ?? []) as UpstreamModel['capabilities'],
    status: raw.status ?? raw.Status ?? '',
    rawMetadata: raw.raw_metadata ?? raw.RawMetadata,
    lastSyncedAt: raw.last_synced_at ?? raw.LastSyncedAt ?? '',
  };
}

function mapUpstreamEvent(raw: RawUpstreamAccountEvent): UpstreamAccountEvent {
  return {
    id: raw.id ?? raw.ID ?? '',
    upstreamAccountId: raw.upstream_account_id ?? raw.UpstreamAccountID ?? '',
    operation: raw.operation ?? raw.Operation ?? '',
    status: raw.status ?? raw.Status ?? '',
    errorClass: raw.error_class ?? raw.ErrorClass,
    message: raw.message ?? raw.Message ?? '',
    latencyMs: raw.latency_ms ?? raw.LatencyMS ?? 0,
    metadata: raw.metadata ?? raw.Metadata,
    createdAt: raw.created_at ?? raw.CreatedAt ?? '',
  };
}
