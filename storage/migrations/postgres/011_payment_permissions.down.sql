UPDATE user_companies uc SET permission_ids = array_remove(uc.permission_ids, p.id)
FROM permissions p WHERE p.slug IN ('add_payment', 'delete_transaction', 'change_payment_status')
  AND p.id = ANY(uc.permission_ids);
DELETE FROM permissions WHERE slug IN ('add_payment', 'delete_transaction', 'change_payment_status');
