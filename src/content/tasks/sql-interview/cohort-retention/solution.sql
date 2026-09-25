WITH months AS (
  SELECT m::date AS cohort
  FROM generate_series(date '2024-01-01', date '2024-04-01', interval '1 month') AS m
),
-- Номер календарного месяца: разность двух таких номеров — «сколько месяцев прошло»
buyer_months AS (
  SELECT DISTINCT user_id,
         extract(year FROM created_at)::int * 12 + extract(month FROM created_at)::int AS ym
  FROM orders
  WHERE status = 'done'
),
cohort_users AS (
  SELECT u.id,
         date_trunc('month', u.signed_up_at)::date AS cohort,
         extract(year FROM u.signed_up_at)::int * 12 + extract(month FROM u.signed_up_at)::int AS ym
  FROM users u
),
flags AS (
  SELECT cu.id, cu.cohort,
         EXISTS (SELECT 1 FROM buyer_months b WHERE b.user_id = cu.id AND b.ym = cu.ym + 1) AS m1,
         EXISTS (SELECT 1 FROM buyer_months b WHERE b.user_id = cu.id AND b.ym = cu.ym + 2) AS m2
  FROM cohort_users cu
)
SELECT m.cohort,
       count(f.id) AS users,
       round(100.0 * count(*) FILTER (WHERE f.m1) / nullif(count(f.id), 0), 1) AS m1_pct,
       round(100.0 * count(*) FILTER (WHERE f.m2) / nullif(count(f.id), 0), 1) AS m2_pct
FROM months m
LEFT JOIN flags f ON f.cohort = m.cohort
GROUP BY m.cohort
ORDER BY m.cohort;
