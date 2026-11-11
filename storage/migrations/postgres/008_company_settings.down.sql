ALTER TABLE users DROP COLUMN IF EXISTS is_super_admin;
DROP TABLE IF EXISTS company_settings;
ALTER TABLE companies
  DROP COLUMN IF EXISTS blocked_by,
  DROP COLUMN IF EXISTS blocked_at,
  DROP COLUMN IF EXISTS is_blocked;
