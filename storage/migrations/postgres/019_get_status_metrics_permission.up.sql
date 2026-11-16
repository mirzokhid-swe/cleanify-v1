-- "Holatlar bo'yicha ko'rsatkichlar" huquqi; get_work_volume_by_day bor
-- xodimlar va rollarga avtomatik beriladi.
WITH ins AS (
  INSERT INTO permissions (slug, name, "group", scope)
  VALUES ('get_status_metrics', 'Holatlar bo''yicha ko''rsatkichlar', 'Statistikalar', 'company')
  ON CONFLICT (slug) DO UPDATE SET name = excluded.name RETURNING id
)
UPDATE user_companies uc SET permission_ids = array_append(uc.permission_ids, ins.id)
FROM ins
WHERE (SELECT id FROM permissions WHERE slug = 'get_work_volume_by_day') = ANY(uc.permission_ids)
  AND NOT (ins.id = ANY(uc.permission_ids));

UPDATE roles r SET permission_ids = array_append(r.permission_ids, p.id)
FROM permissions p
WHERE p.slug = 'get_status_metrics'
  AND (SELECT id FROM permissions WHERE slug = 'get_work_volume_by_day') = ANY(r.permission_ids)
  AND NOT (p.id = ANY(r.permission_ids));
