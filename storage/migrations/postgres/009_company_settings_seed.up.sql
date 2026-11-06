-- Boshlang'ich qiymatlar: hozir kodda company_id bo'yicha yozilgan farqlar
-- (front feature-flags.js, Orders/Item.jsx, orderStatusLock.js, Volume/Item.jsx,
-- Laravel TelegramController $prod_company_id).
INSERT INTO company_settings (company_id, key, value) VALUES
  ('32eea25d-4412-4288-88ce-5ddd5db7612d', 'order_card_code',       'true'::jsonb),     -- Pokiza namangan
  ('5ded78ed-35b0-4b05-b9e7-be1d988b0740', 'order_card_code',       'true'::jsonb),     -- POKIZA NUR
  ('82e1054c-48ef-4ed5-9e6b-9d3fbbf00477', 'lock_delivered_status', 'true'::jsonb),     -- Yangi Dunyo - Qoratepa
  ('82e1054c-48ef-4ed5-9e6b-9d3fbbf00477', 'kpi_enabled',           'true'::jsonb),
  ('82e1054c-48ef-4ed5-9e6b-9d3fbbf00477', 'volume_mode',           '"worker"'::jsonb),
  ('82e1054c-48ef-4ed5-9e6b-9d3fbbf00478', 'volume_mode',           '"worker"'::jsonb), -- Yangi Dunyo - Asnam
  ('cd3ef1fd-4104-4774-9247-f1a5d7e3ab1b', 'volume_mode',           '"worker"'::jsonb), -- Mega Chistka
  ('2de68d2e-14cb-4ff0-901d-f165aa53fab5', 'volume_skip_countable', 'true'::jsonb)      -- Musaffo
ON CONFLICT DO NOTHING;

UPDATE companies SET is_blocked = true, blocked_at = now()
WHERE id IN (
  '2de68d2e-14cb-4ff0-901d-f165aa53fab5', -- Musaffo
  '648306f6-e670-4019-8fd3-21bd813d40a7', -- Ishonch Namangan
  '36c6dcd3-8097-40a9-8817-a5b42ad49203'  -- Ko'rkam
) AND NOT is_blocked;
