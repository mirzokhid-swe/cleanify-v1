-- kpi_stats_mv qayta yaratiladi: o'chirilgan to'lovlar (deleted_at) hisobga
-- kirmaydi. Manba: Laravel 2026_03_14_170000_create_kpi_stats_materialized_view.
DROP MATERIALIZED VIEW IF EXISTS kpi_stats_mv;

CREATE MATERIALIZED VIEW kpi_stats_mv AS
WITH delivered AS (
    SELECT DISTINCT ON (historyable_id)
        historyable_id AS order_id,
        user_id        AS delivered_by,
        created_at::date AS delivered_date
    FROM status_change_histories
    WHERE historyable_type = 'orders'
      AND status = 6
    ORDER BY historyable_id, created_at DESC
),
accepted AS (
    SELECT DISTINCT ON (historyable_id)
        historyable_id AS order_id,
        user_id        AS accepted_by
    FROM status_change_histories
    WHERE historyable_type = 'orders'
      AND status = 2
    ORDER BY historyable_id, created_at DESC
),
order_kpi AS (
    SELECT
        o.company_id,
        d.delivered_date,
        d.order_id,
        d.delivered_by,
        a.accepted_by
    FROM delivered d
    INNER JOIN orders o ON o.id = d.order_id
    LEFT JOIN accepted a ON a.order_id = d.order_id
),
order_transactions AS (
    SELECT
        payer_id::int AS order_id,
        SUM(amount)   AS amount
    FROM transactions
    WHERE payer_type = 'orders'
      AND receiver_type = 'users'
      AND deleted_at IS NULL
    GROUP BY payer_id::int
),
delivered_agg AS (
    SELECT
        ok.company_id,
        ok.delivered_by         AS user_id,
        ok.delivered_date       AS stat_date,
        COUNT(*)                AS delivered_count,
        COALESCE(SUM(ot.amount), 0) AS delivered_sum,
        ARRAY_AGG(ok.order_id)  AS delivered_ids
    FROM order_kpi ok
    LEFT JOIN order_transactions ot
        ON ot.order_id = ok.order_id
    GROUP BY ok.company_id, ok.delivered_by, ok.delivered_date
),
accepted_agg AS (
    SELECT
        ok.company_id,
        ok.accepted_by          AS user_id,
        ok.delivered_date       AS stat_date,
        COUNT(*)                AS accepted_count,
        COALESCE(SUM(ot.amount), 0) AS accepted_sum,
        ARRAY_AGG(ok.order_id)  AS accepted_ids
    FROM order_kpi ok
    LEFT JOIN order_transactions ot
        ON ot.order_id = ok.order_id
    WHERE ok.accepted_by IS NOT NULL
    GROUP BY ok.company_id, ok.accepted_by, ok.delivered_date
),
all_users AS (
    SELECT company_id, user_id, stat_date FROM delivered_agg
    UNION
    SELECT company_id, user_id, stat_date FROM accepted_agg
)
SELECT
    au.company_id,
    au.user_id,
    au.stat_date                            AS date,
    u.fullname,
    COALESCE(da.delivered_count, 0)         AS delivered_count,
    COALESCE(da.delivered_sum,   0)         AS delivered_sum,
    COALESCE(da.delivered_ids,   '{}')      AS delivered_ids,
    COALESCE(aa.accepted_count,  0)         AS accepted_count,
    COALESCE(aa.accepted_sum,    0)         AS accepted_sum,
    COALESCE(aa.accepted_ids,    '{}')      AS accepted_ids
FROM all_users au
LEFT JOIN users u
    ON u.id = au.user_id
LEFT JOIN delivered_agg da
    ON da.company_id = au.company_id
   AND da.user_id    = au.user_id
   AND da.stat_date  = au.stat_date
LEFT JOIN accepted_agg aa
    ON aa.company_id = au.company_id
   AND aa.user_id    = au.user_id
   AND aa.stat_date  = au.stat_date
ORDER BY au.stat_date DESC, au.company_id, u.fullname;

CREATE UNIQUE INDEX kpi_stats_mv_unique_idx ON kpi_stats_mv (company_id, user_id, date);
CREATE INDEX kpi_stats_mv_company_date_idx ON kpi_stats_mv (company_id, date);
