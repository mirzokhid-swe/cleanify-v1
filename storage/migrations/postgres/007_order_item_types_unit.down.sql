DROP INDEX IF EXISTS idx_order_item_types_company;
ALTER TABLE order_item_types DROP COLUMN IF EXISTS unit;
