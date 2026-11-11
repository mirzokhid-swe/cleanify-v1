ALTER TABLE companies
  ADD COLUMN IF NOT EXISTS is_blocked boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS blocked_at timestamptz,
  ADD COLUMN IF NOT EXISTS blocked_by bigint;

CREATE TABLE IF NOT EXISTS company_settings (
  company_id uuid NOT NULL REFERENCES companies(id),
  key        varchar(60) NOT NULL,
  value      jsonb NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by bigint,
  PRIMARY KEY (company_id, key)
);

ALTER TABLE users ADD COLUMN IF NOT EXISTS is_super_admin boolean NOT NULL DEFAULT false;
