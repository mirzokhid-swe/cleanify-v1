-- Yangi huquqlar (guruh "Buyurtmalar"). Boshlang'ich: add_payment —
-- get_orders_list (8) bor hammaga; qolgan ikkitasi — delete_order (37) borlarga.
WITH ins AS (
  INSERT INTO permissions (slug, name, "group", scope)
  VALUES ('add_payment', 'To''lov kiritish', 'Buyurtmalar', 'company')
  ON CONFLICT (slug) DO UPDATE SET name = excluded.name RETURNING id
)
UPDATE user_companies uc SET permission_ids = array_append(uc.permission_ids, ins.id)
FROM ins WHERE 8 = ANY(uc.permission_ids) AND NOT (ins.id = ANY(uc.permission_ids));

WITH ins AS (
  INSERT INTO permissions (slug, name, "group", scope)
  VALUES ('delete_transaction', 'To''lovni o''chirish', 'Buyurtmalar', 'company')
  ON CONFLICT (slug) DO UPDATE SET name = excluded.name RETURNING id
)
UPDATE user_companies uc SET permission_ids = array_append(uc.permission_ids, ins.id)
FROM ins WHERE 37 = ANY(uc.permission_ids) AND NOT (ins.id = ANY(uc.permission_ids));

WITH ins AS (
  INSERT INTO permissions (slug, name, "group", scope)
  VALUES ('change_payment_status', 'To''lov holatini o''zgartirish', 'Buyurtmalar', 'company')
  ON CONFLICT (slug) DO UPDATE SET name = excluded.name RETURNING id
)
UPDATE user_companies uc SET permission_ids = array_append(uc.permission_ids, ins.id)
FROM ins WHERE 37 = ANY(uc.permission_ids) AND NOT (ins.id = ANY(uc.permission_ids));
