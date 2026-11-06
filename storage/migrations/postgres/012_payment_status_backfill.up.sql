-- Ma'lumotni ko'chirish (010, 011, 013 dan keyin; bir necha soniya). Bitta
-- tranzaksiyada: yarim yo'lda xato bo'lsa hech narsa o'zgarmaydi.
BEGIN;

-- 1) Yangi Dunyo - Qoratepa: "Qarzdor" (11) holati olib tashlanadi — qarzdor
--    buyurtma = topshirilgan + to'lov holati to'liq emas.
UPDATE orders SET status = 6 WHERE status = 11;
DELETE FROM order_statuses WHERE number = 11;
UPDATE user_companies uc SET permission_ids = array_remove(uc.permission_ids, p.id)
FROM permissions p WHERE p.slug = 'change_status_to_11' AND p.id = ANY(uc.permission_ids);
DELETE FROM permissions WHERE slug = 'change_status_to_11';

-- 2) Eski topshirilgan buyurtmalar "To'langan" deb olinadi (eski qarzlar
--    kuzatilmaydi). Sifat'da to'lovlar minglarda yozilgani uchun chegara bugun.
UPDATE orders
SET payment_status = 3, payment_status_manual = true
WHERE status = 6
  AND created_at < CASE WHEN company_id = '3fd0c9ec-66c3-4010-966e-378cf0db5b3c'
                        THEN now() ELSE '2026-09-01'::timestamptz END;

-- 3) Qolganlari funksiya bilan hisoblanadi.
SELECT count(order_payment_status_recalc(id)) AS recalculated FROM orders WHERE NOT payment_status_manual;

COMMIT;
