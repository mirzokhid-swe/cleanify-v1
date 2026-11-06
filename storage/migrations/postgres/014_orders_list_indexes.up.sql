CREATE INDEX IF NOT EXISTS idx_orders_company_status ON orders (company_id, status);
CREATE INDEX IF NOT EXISTS idx_orders_company_created ON orders (company_id, created_at DESC);
-- idx_order_items_order va idx_transactions_order 010 da.
