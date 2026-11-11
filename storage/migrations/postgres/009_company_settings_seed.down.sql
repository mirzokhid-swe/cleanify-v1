DELETE FROM company_settings
WHERE key IN ('order_card_code', 'lock_delivered_status', 'kpi_enabled', 'volume_mode', 'volume_skip_countable')
  AND company_id IN (
    '32eea25d-4412-4288-88ce-5ddd5db7612d', '5ded78ed-35b0-4b05-b9e7-be1d988b0740',
    '82e1054c-48ef-4ed5-9e6b-9d3fbbf00477', '82e1054c-48ef-4ed5-9e6b-9d3fbbf00478',
    'cd3ef1fd-4104-4774-9247-f1a5d7e3ab1b', '2de68d2e-14cb-4ff0-901d-f165aa53fab5');

UPDATE companies SET is_blocked = false, blocked_at = NULL, blocked_by = NULL
WHERE id IN ('2de68d2e-14cb-4ff0-901d-f165aa53fab5', '648306f6-e670-4019-8fd3-21bd813d40a7', '36c6dcd3-8097-40a9-8817-a5b42ad49203');
