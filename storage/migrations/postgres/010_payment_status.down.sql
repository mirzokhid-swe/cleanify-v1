DROP FUNCTION IF EXISTS order_payment_status_recalc(integer);
DROP INDEX IF EXISTS idx_order_items_order;
DROP INDEX IF EXISTS idx_transactions_order;
ALTER TABLE orders DROP COLUMN IF EXISTS slots, DROP COLUMN IF EXISTS payment_status_manual;
ALTER TABLE transactions DROP COLUMN IF EXISTS deleted_by, DROP COLUMN IF EXISTS deleted_at;
