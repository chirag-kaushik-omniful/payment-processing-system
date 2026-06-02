ALTER TABLE payments ADD COLUMN IF NOT EXISTS amount_cents BIGINT;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS base_currency TEXT DEFAULT 'USD';
ALTER TABLE payments ADD COLUMN IF NOT EXISTS fx_rate NUMERIC(18,8) DEFAULT 1;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS payment_method_id UUID;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS hold_id UUID;

CREATE TABLE IF NOT EXISTS payment_methods (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  provider TEXT NOT NULL,
  token TEXT NOT NULL,
  last_four TEXT,
  brand TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS payment_holds (
  id UUID PRIMARY KEY,
  payment_id UUID REFERENCES payments(id),
  user_id UUID NOT NULL,
  amount NUMERIC(18,2) NOT NULL,
  currency TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'HELD',
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  captured_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS audit_events (
  id UUID PRIMARY KEY,
  actor_id TEXT,
  action TEXT NOT NULL,
  resource_type TEXT NOT NULL,
  resource_id TEXT NOT NULL,
  metadata JSONB,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS disputes (
  id UUID PRIMARY KEY,
  payment_id UUID REFERENCES payments(id),
  provider TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'OPEN',
  reason TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_events_resource ON audit_events(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_disputes_payment ON disputes(payment_id);
