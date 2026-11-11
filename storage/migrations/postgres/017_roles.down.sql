-- Huquqlar (delete_user, manage_roles), tozalangan user_companies qatorlari va
-- DROP qilingan eski jadvallar qaytmaydi.
DROP INDEX IF EXISTS user_companies_user_company_key;
ALTER TABLE user_companies DROP COLUMN IF EXISTS role_id;
DROP TABLE IF EXISTS roles;
