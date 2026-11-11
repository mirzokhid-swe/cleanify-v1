ALTER TABLE order_item_types
  ADD COLUMN IF NOT EXISTS unit varchar(10) NOT NULL DEFAULT 'area'
  CHECK (unit IN ('area', 'piece', 'length', 'weight'));

UPDATE order_item_types SET unit = 'piece' WHERE is_countable;

CREATE INDEX IF NOT EXISTS idx_order_item_types_company ON order_item_types (company_id);

-- GET /order-item-type endi get_order_item_type (27) huquqini talab qiladi;
-- buyum qo'shadigan/tahrirlaydigan (20, 21) yoki buyurtmalar ro'yxatini
-- ko'radigan (8) har bir xodimga 27 beriladi, aks holda buyurtma sahifasida
-- turlar ro'yxati ochilmaydi.
UPDATE user_companies
SET permission_ids = array_append(permission_ids, 27)
WHERE (8 = ANY(permission_ids) OR 20 = ANY(permission_ids) OR 21 = ANY(permission_ids))
  AND NOT 27 = ANY(permission_ids);
