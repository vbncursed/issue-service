CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS issuer_keys (
  key_id TEXT PRIMARY KEY,
  alg TEXT NOT NULL,
  public_key BYTEA NOT NULL,
  private_key BYTEA NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('active','retired')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS passes (
  id UUID PRIMARY KEY,
  org_id UUID NOT NULL,
  policy_id TEXT NOT NULL,
  subject_name TEXT NOT NULL,
  zone_id TEXT NOT NULL,
  nbf TIMESTAMPTZ NOT NULL,
  exp TIMESTAMPTZ NOT NULL,
  one_time BOOLEAN NOT NULL DEFAULT TRUE,
  issuer_key_id TEXT NOT NULL REFERENCES issuer_keys(key_id),
  signature BYTEA NOT NULL,
  payload BYTEA NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('Active','Revoked','Expired')) DEFAULT 'Active'
);

CREATE INDEX IF NOT EXISTS idx_passes_status ON passes(status);
CREATE INDEX IF NOT EXISTS idx_passes_exp ON passes(exp);
CREATE INDEX IF NOT EXISTS idx_passes_org ON passes(org_id);

CREATE TABLE IF NOT EXISTS pickup_tokens (
  token TEXT PRIMARY KEY,
  pass_id UUID NOT NULL REFERENCES passes(id) ON DELETE CASCADE,
  ttl_expires_at TIMESTAMPTZ NOT NULL,
  used_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_pickup_expires ON pickup_tokens(ttl_expires_at);


