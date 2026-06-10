export type ThemeMode = 'dark' | 'light';

export interface AdminUser {
  id: string;
  email: string;
  role: 'owner' | 'admin' | 'developer' | 'viewer';
  status: 'active' | 'inactive' | 'blocked';
}

export type PageId =
  | 'overview'
  | 'sites'
  | 'models'
  | 'routing'
  | 'checkin'
  | 'quota'
  | 'testing'
  | 'users'
  | 'apiKeys'
  | 'logs'
  | 'settings';

export type SiteStatus = 'normal' | 'warning' | 'failed' | 'maintenance' | 'offline';
export type ModelStatus = 'normal' | 'partial' | 'unavailable';
export type CheckinStatus = 'checked' | 'unchecked' | 'disabled';
export type TestStatus = 'success' | 'partial' | 'failed';
export type UserRole = 'super_admin' | 'admin' | 'developer' | 'viewer';
export type UserStatus = 'active' | 'inactive' | 'blocked';
export type ApiKeyStatus = 'active' | 'expiring' | 'blocked' | 'expired' | 'unused';
export type ApiKeyScope = 'Chat' | 'Embedding' | 'Vision' | 'Admin';
export type AlertSeverity = 'info' | 'warning' | 'danger' | 'success';
export type SiteType = 'OpenAI 兼容' | 'Claude' | 'Gemini' | 'Azure OpenAI';
export type Capability = 'stream' | 'function' | 'vision' | 'embedding';

export interface ChartPoint {
  label: string;
  value: number;
}

export interface Metric {
  label: string;
  value: string;
  delta: string;
  tone: AlertSeverity;
}

export interface Site {
  id: string;
  code: string;
  name: string;
  url: string;
  type: SiteType;
  region: string;
  flag: string;
  status: SiteStatus;
  models: string[];
  latencyMs?: number;
  balanceUsd: number;
  checkinStatus: CheckinStatus;
  todayCalls: number;
  successRate: number;
  healthScore: number;
  lastChecked: string;
  note: string;
}

export interface ModelInfo {
  id: string;
  name: string;
  provider: string;
  iconText: string;
  status: ModelStatus;
  kind: string[];
  availableSites: number;
  totalSites: number;
  recommendedSiteId: string;
  minLatencyMs?: number;
  successRate: number;
  quotaUsage: number;
  quotaLabel: string;
  routingMode: string;
  capabilities: Capability[];
  trend: number[];
  score: number;
  todayCalls: number;
  createdAt: string;
}

export interface AlertItem {
  id: string;
  severity: AlertSeverity;
  title: string;
  description: string;
  time: string;
}

export interface RoutingCandidate {
  id: string;
  rank: number;
  siteId: string;
  manualWeight: number;
  healthScore: number;
  successRate: number;
  latencyMs?: number;
  load: number;
  circuitState: 'closed' | 'open' | 'cooldown';
  score: number;
}

export interface RouteHistory {
  id: string;
  siteName: string;
  result: 'success' | 'failed';
  score?: number;
  time: string;
}

export interface ScoreBreakdown {
  label: string;
  value: number;
  max: number;
  tone: AlertSeverity;
}

export interface QuotaRecord {
  id: string;
  siteId: string;
  mode: '自动签到' | '手动签到' | '维护中';
  model: string;
  lastCheckin: string;
  rewardUsd?: number;
  currentUsd: number;
  status: CheckinStatus;
}

export interface TestTemplate {
  id: string;
  name: string;
  model: string;
  scope: string;
  time: string;
  tone: AlertSeverity;
}

export interface TestResultRow {
  id: string;
  siteId: string;
  status: TestStatus;
  latencyMs: number;
  tokens: number;
  error?: string;
  supports: Record<Capability, boolean>;
  testedAt: string;
}

export interface UserActivity {
  id: string;
  type: 'login' | 'key' | 'role' | 'quota';
  title: string;
  description: string;
  time: string;
}

export interface UserRecord {
  id: string;
  name: string;
  email: string;
  avatarText: string;
  role: UserRole;
  status: UserStatus;
  organization: string;
  note: string;
  availableModels: string[];
  apiKeyCount: number;
  monthlyQuotaUsd: number;
  usedQuotaUsd: number;
  rateLimit: string;
  joinedAt: string;
  lastLogin: string;
  activity: UserActivity[];
}

export interface ApiKeyCall {
  id: string;
  endpoint: string;
  statusCode: number;
  time: string;
}

export interface ApiKeyRecord {
  id: string;
  name: string;
  ownerName: string;
  ownerEmail: string;
  prefix: string;
  maskedKey: string;
  scopes: ApiKeyScope[];
  allowedModels: string[];
  rateLimit: string;
  dailyLimit: string;
  ipWhitelist: string;
  createdAt: string;
  expiresAt: string;
  lastUsed: string;
  status: ApiKeyStatus;
  usageTokens: number;
  usagePercent: number;
  recentCalls: ApiKeyCall[];
}
