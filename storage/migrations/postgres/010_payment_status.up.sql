ALTER TABLE transactions
  ADD COLUMN IF NOT EXISTS deleted_at timestamptz,
  ADD COLUMN IF NOT EXISTS deleted_by bigint;

ALTER TABLE orders
  ADD COLUMN IF NOT EXISTS payment_status_manual boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS slots text[];

CREATE INDEX IF NOT EXISTS idx_transactions_order ON transactions (payer_id) WHERE payer_type = 'orders';
CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items (order_id);

-- To'lov holati qoidasi bitta joyda. Go uni to'lov va buyum o'zgarishlaridan
-- keyin chaqiradi, backfill (012) ham shu funksiyani ishlatadi. Trigger yo'q.
-- Summalar butun so'mgacha yaxlitlanadi: buyum o'lchamlari float32 orqali
-- kelgani uchun jami 175599.9946 kabi chiqadi, xodim esa 175 600 to'laydi.
-- NULL: buyum ham, to'lov ham yo'q (ekranda "To'lanmagan"); 1 To'lanmagan;
-- 2 Qisman; 3 To'langan (qolgan <= payment_tolerance). payment_status_manual
-- bo'lsa o'zgartirmaydi va NULL qaytaradi.
CREATE OR REPLACE FUNCTION order_payment_status_recalc(p_order_id integer) RETURNS integer
LANGUAGE plpgsql AS $$
DECLARE
  v_company uuid; v_manual boolean; v_total numeric; v_paid numeric; v_tol numeric; v_status integer;
BEGIN
  SELECT company_id, payment_status_manual INTO v_company, v_manual FROM orders WHERE id = p_order_id;
  IF NOT FOUND OR v_manual THEN RETURN NULL; END IF;
  SELECT ROUND(COALESCE(SUM(price * width * height), 0)::numeric) INTO v_total FROM order_items WHERE order_id = p_order_id;
  SELECT ROUND(COALESCE(SUM(amount), 0)::numeric) INTO v_paid FROM transactions
    WHERE payer_type = 'orders' AND payer_id = p_order_id AND deleted_at IS NULL;
  SELECT (value #>> '{}')::numeric INTO v_tol FROM company_settings
    WHERE company_id = v_company AND key = 'payment_tolerance';
  IF v_total = 0 AND v_paid = 0 THEN v_status := NULL;
  ELSIF v_paid = 0 THEN v_status := 1;
  ELSIF v_total - v_paid <= COALESCE(v_tol, 0) THEN v_status := 3;
  ELSE v_status := 2;
  END IF;
  UPDATE orders SET payment_status = v_status WHERE id = p_order_id;
  RETURN v_status;
END $$;
