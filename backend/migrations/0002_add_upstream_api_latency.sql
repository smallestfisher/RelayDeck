ALTER TABLE upstream_account_status
  ADD COLUMN IF NOT EXISTS api_latency_ms INTEGER NOT NULL DEFAULT 0;
