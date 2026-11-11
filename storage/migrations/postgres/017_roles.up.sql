-- Rollar (hybrid): user_companies.role_id bo'lsa roles.permission_ids, bo'lmasa o'z permission_ids.

-- Eski, ishlatilmaydigan jadvallar
DROP TABLE IF EXISTS role_and_permissions, user_permissions;

CREATE TABLE IF NOT EXISTS roles (
  id             bigserial PRIMARY KEY,
  company_id     uuid NOT NULL REFERENCES companies(id),
  name           varchar(100) NOT NULL,
  permission_ids integer[] NOT NULL DEFAULT '{}',
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now(),
  UNIQUE (company_id, name)
);

-- user_companies tozalash: takrorlardan eng katta id qoladi
DELETE FROM user_companies a USING user_companies b
WHERE a.user_id = b.user_id AND a.company_id = b.company_id AND a.id < b.id;

-- egasiz qatorlar o'chadi
DELETE FROM user_companies uc
WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = uc.user_id);

-- users.company_id bor, lekin user_companies da qatori yo'q tirik xodimlarga qator
INSERT INTO user_companies (permission_ids, company_id, user_id, is_courier)
SELECT '{}', u.company_id::text, u.id, false
FROM users u
WHERE u.company_id IS NOT NULL AND u.deleted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM user_companies uc WHERE uc.user_id = u.id);

UPDATE user_companies SET permission_ids = '{}' WHERE permission_ids IS NULL;

ALTER TABLE user_companies ADD COLUMN IF NOT EXISTS role_id bigint REFERENCES roles(id);
CREATE UNIQUE INDEX IF NOT EXISTS user_companies_user_company_key ON user_companies (user_id, company_id);

-- Yangi huquqlar (guruh "Foydalanuvchilar"): user_create yoki edit_user bor hammaga
WITH ins AS (
  INSERT INTO permissions (slug, name, "group", scope)
  VALUES ('delete_user', 'O''chirish', 'Foydalanuvchilar', 'company')
  ON CONFLICT (slug) DO UPDATE SET name = excluded.name RETURNING id
)
UPDATE user_companies uc SET permission_ids = array_append(uc.permission_ids, ins.id)
FROM ins
WHERE (
    (SELECT id FROM permissions WHERE slug = 'user_create') = ANY(uc.permission_ids)
    OR (SELECT id FROM permissions WHERE slug = 'edit_user') = ANY(uc.permission_ids)
  ) AND NOT (ins.id = ANY(uc.permission_ids));

WITH ins AS (
  INSERT INTO permissions (slug, name, "group", scope)
  VALUES ('manage_roles', 'Rollarni boshqarish', 'Foydalanuvchilar', 'company')
  ON CONFLICT (slug) DO UPDATE SET name = excluded.name RETURNING id
)
UPDATE user_companies uc SET permission_ids = array_append(uc.permission_ids, ins.id)
FROM ins
WHERE (
    (SELECT id FROM permissions WHERE slug = 'user_create') = ANY(uc.permission_ids)
    OR (SELECT id FROM permissions WHERE slug = 'edit_user') = ANY(uc.permission_ids)
  ) AND NOT (ins.id = ANY(uc.permission_ids));
