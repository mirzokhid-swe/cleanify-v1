-- Foydalanuvchilarni soft delete: deleted_at / deleted_by.
-- Telefon unikalligi faqat tiriklar orasida (raqam qayta ishlatilishi uchun).
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at timestamptz;
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_by bigint;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_phone_key;
CREATE UNIQUE INDEX IF NOT EXISTS users_phone_active_key ON users (phone) WHERE deleted_at IS NULL;
