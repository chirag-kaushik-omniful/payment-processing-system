CREATE TABLE IF NOT EXISTS wallet_holds (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES wallets(user_id),
  amount NUMERIC(18, 2) NOT NULL,
  status TEXT NOT NULL DEFAULT 'ACTIVE',
  payment_id UUID,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  released_at TIMESTAMP
);
