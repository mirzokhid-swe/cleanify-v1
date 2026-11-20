CREATE TABLE IF NOT EXISTS employee_attendance (
    id BIGSERIAL PRIMARY KEY,
    company_id UUID NOT NULL REFERENCES companies(id),
    employee_id INT NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    status SMALLINT NOT NULL CHECK (status IN (0, 1, 2)),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (employee_id, date)
);

CREATE INDEX IF NOT EXISTS employee_attendance_company_date_idx
    ON employee_attendance (company_id, date);

-- Eski JSON dan faqat 1 (yarim) va 2 (to'liq). 0 ko'chirilmaydi: eski tizimda u "belgilanmagan" edi.
INSERT INTO employee_attendance (company_id, employee_id, date, status)
SELECT a.company_id, e.id, a.date, max((r ->> 'work_schedule')::int)
FROM attendance a
CROSS JOIN LATERAL jsonb_array_elements(
    CASE WHEN jsonb_typeof(a.employees) = 'array' THEN a.employees ELSE '[]'::jsonb END
) r
JOIN employees e
    ON e.id = (r ->> 'employee_id')::int AND e.company_id = a.company_id
WHERE a.date IS NOT NULL
  AND (r ->> 'work_schedule') IN ('1', '2')
  AND (r ->> 'employee_id') ~ '^[0-9]+$'
GROUP BY a.company_id, e.id, a.date
ON CONFLICT (employee_id, date) DO NOTHING;
