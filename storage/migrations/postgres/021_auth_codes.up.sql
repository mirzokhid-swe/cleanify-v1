CREATE TABLE IF NOT EXISTS auth_codes (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL,
    code varchar(6) NOT NULL,
    expires_at timestamptz NOT NULL,
    used_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS auth_codes_active_code_idx ON auth_codes (code) WHERE used_at IS NULL;
