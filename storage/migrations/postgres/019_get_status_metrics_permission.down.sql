UPDATE user_companies uc SET permission_ids = array_remove(uc.permission_ids, p.id)
FROM permissions p WHERE p.slug = 'get_status_metrics';
UPDATE roles r SET permission_ids = array_remove(r.permission_ids, p.id)
FROM permissions p WHERE p.slug = 'get_status_metrics';
DELETE FROM permissions WHERE slug = 'get_status_metrics';
