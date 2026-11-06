ALTER TABLE order_item_types
  ADD COLUMN IF NOT EXISTS unit varchar(10) NOT NULL DEFAULT 'area'
  CHECK (unit IN ('area', 'piece', 'length', 'weight'));

UPDATE order_item_types SET unit = 'piece' WHERE is_countable;

CREATE INDEX IF NOT EXISTS idx_order_item_types_company ON order_item_types (company_id);
