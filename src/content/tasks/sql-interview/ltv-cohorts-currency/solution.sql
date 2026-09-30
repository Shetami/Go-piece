WITH cohorts AS (
  SELECT c::date AS cohort
  FROM generate_series(date '2024-01-01', date '2024-03-01', interval '1 month') AS c
),
members AS (
  SELECT id, signup_on, date_trunc('month', signup_on)::date AS cohort
  FROM users
  WHERE NOT is_test
),
revenue AS (
  SELECT m.id, m.cohort,
         sum(o.amount * CASE WHEN o.currency = 'USD' THEN 1 ELSE fx.rate_to_usd END) AS usd
  FROM members m
  JOIN orders o ON o.user_id = m.id
               AND o.status = 'paid'
               AND o.created_on >= m.signup_on
               AND o.created_on <  m.signup_on + 30
  LEFT JOIN fx_monthly fx ON fx.currency = o.currency
                         AND fx.month = date_trunc('month', o.created_on)::date
  GROUP BY m.id, m.cohort
)
SELECT c.cohort,
       count(m.id)                       AS users,
       count(r.id)                       AS paying_users,
       round(coalesce(sum(r.usd), 0), 2) AS revenue_usd,
       round(sum(r.usd) / nullif(count(m.id), 0), 2) AS ltv30_usd
FROM cohorts c
LEFT JOIN members m ON m.cohort = c.cohort
LEFT JOIN revenue r ON r.id = m.id
GROUP BY c.cohort
ORDER BY c.cohort;
