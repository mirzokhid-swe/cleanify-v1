-- Diqqat: o'chirilgan va tirik xodimda bir xil telefon bo'lsa, users_phone_key
-- qaytarilmaydi (muvaffaqiyatsiz bo'ladi) — oldin takrorlarni qo'lda hal qiling.
DROP INDEX IF EXISTS users_phone_active_key;
ALTER TABLE users ADD CONSTRAINT users_phone_key UNIQUE (phone);
ALTER TABLE users DROP COLUMN IF EXISTS deleted_by;
ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;
