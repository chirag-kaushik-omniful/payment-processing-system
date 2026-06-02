CREATE TABLE IF NOT EXISTS wallets (
  user_id UUID PRIMARY KEY,
  balance NUMERIC(18, 2) NOT NULL DEFAULT 0,
  version INT NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wallets_updated_at ON wallets(updated_at);
