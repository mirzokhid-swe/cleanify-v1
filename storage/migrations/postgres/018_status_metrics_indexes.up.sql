CREATE INDEX IF NOT EXISTS idx_sch_orders_status_created
  ON status_change_histories (status, created_at) WHERE historyable_type = 'orders';
CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items (order_id);
