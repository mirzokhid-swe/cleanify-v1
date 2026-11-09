-- Buyumning o'lchov birligi: area (m²), piece (dona), length (metr), weight (kg).
-- Yangi buyumlarga Go turdan nusxa qiladi; tur o'chirilsa yoki birligi o'zgarsa ham
-- eski buyum o'z birligini saqlaydi.
ALTER TABLE order_items
  ADD COLUMN IF NOT EXISTS unit varchar(10) NOT NULL DEFAULT 'area'
  CHECK (unit IN ('area', 'piece', 'length', 'weight'));

-- Eski buyumlar: buyumning o'z is_countable qiymati asos (tur keyin o'zgargan bo'lishi
-- mumkin: o'lcham bilan saqlangan buyum turi keyin "dona" qilingan va aksincha).
UPDATE order_items
SET unit = CASE WHEN is_countable THEN 'piece' ELSE 'area' END
WHERE unit <> CASE WHEN is_countable THEN 'piece' ELSE 'area' END
  AND unit NOT IN ('length', 'weight');

-- Donalik bo'lmagan buyum, turi metr yoki kg — o'sha birlik.
UPDATE order_items oi SET unit = t.unit
FROM order_item_types t
WHERE oi.order_item_type_id = t.id
  AND NOT oi.is_countable
  AND t.unit IN ('length', 'weight')
  AND oi.unit <> t.unit;
