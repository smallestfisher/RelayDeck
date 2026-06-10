CREATE TABLE users (
  id UUID PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE api_keys (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id),
  name TEXT NOT NULL,
  key_prefix TEXT NOT NULL,
  key_hash TEXT NOT NULL UNIQUE,
  scopes TEXT[] NOT NULL DEFAULT '{}',
  allowed_models TEXT[] NOT NULL DEFAULT '{}',
  rpm_limit INTEGER NOT NULL DEFAULT 60,
  tpm_limit INTEGER NOT NULL DEFAULT 10000,
  monthly_quota_tokens BIGINT,
  status TEXT NOT NULL,
  expires_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE upstream_sites (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  base_url TEXT NOT NULL,
  site_type TEXT NOT NULL,
  region TEXT,
  status TEXT NOT NULL,
  weight NUMERIC NOT NULL DEFAULT 50,
  timeout_ms INTEGER NOT NULL DEFAULT 30000,
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE upstream_credentials (
  id UUID PRIMARY KEY,
  site_id UUID NOT NULL REFERENCES upstream_sites(id),
  name TEXT NOT NULL,
  encrypted_secret BYTEA NOT NULL,
  key_prefix TEXT,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE models (
  id TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  capabilities TEXT[] NOT NULL DEFAULT '{}',
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE site_models (
  id UUID PRIMARY KEY,
  site_id UUID NOT NULL REFERENCES upstream_sites(id),
  model_id TEXT NOT NULL REFERENCES models(id),
  upstream_model TEXT NOT NULL,
  endpoint_types TEXT[] NOT NULL DEFAULT '{}',
  capabilities TEXT[] NOT NULL DEFAULT '{}',
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(site_id, model_id, upstream_model)
);

CREATE TABLE routing_policies (
  id UUID PRIMARY KEY,
  model_id TEXT REFERENCES models(id),
  mode TEXT NOT NULL,
  minimum_health_score NUMERIC NOT NULL DEFAULT 70,
  retry_count INTEGER NOT NULL DEFAULT 2,
  circuit_cooldown_seconds INTEGER NOT NULL DEFAULT 60,
  config JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE request_logs (
  id UUID PRIMARY KEY,
  api_key_id UUID REFERENCES api_keys(id),
  user_id UUID REFERENCES users(id),
  endpoint_type TEXT NOT NULL,
  model_id TEXT,
  selected_site_id UUID REFERENCES upstream_sites(id),
  status_code INTEGER,
  duration_ms INTEGER,
  prompt_tokens INTEGER,
  completion_tokens INTEGER,
  total_tokens INTEGER,
  stream BOOLEAN NOT NULL DEFAULT false,
  error_code TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE request_attempts (
  id UUID PRIMARY KEY,
  request_log_id UUID REFERENCES request_logs(id),
  site_id UUID REFERENCES upstream_sites(id),
  upstream_model TEXT,
  status_code INTEGER,
  duration_ms INTEGER,
  error_code TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE upstream_health_checks (
  id UUID PRIMARY KEY,
  site_id UUID NOT NULL REFERENCES upstream_sites(id),
  status TEXT NOT NULL,
  latency_ms INTEGER,
  error_code TEXT,
  capabilities TEXT[] NOT NULL DEFAULT '{}',
  checked_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE usage_daily (
  day DATE NOT NULL,
  user_id UUID REFERENCES users(id),
  api_key_id UUID REFERENCES api_keys(id),
  model_id TEXT REFERENCES models(id),
  site_id UUID REFERENCES upstream_sites(id),
  request_count BIGINT NOT NULL DEFAULT 0,
  failed_count BIGINT NOT NULL DEFAULT 0,
  prompt_tokens BIGINT NOT NULL DEFAULT 0,
  completion_tokens BIGINT NOT NULL DEFAULT 0,
  total_tokens BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY(day, user_id, api_key_id, model_id, site_id)
);
