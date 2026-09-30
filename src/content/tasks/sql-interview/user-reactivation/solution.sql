WITH activity AS (
  SELECT DISTINCT e.user_id, date_trunc('month', e.ts)::date AS month
  FROM events e
  JOIN users u ON u.id = e.user_id
  WHERE NOT u.is_test
),
first_seen AS (
  SELECT user_id, min(month) AS first_month
  FROM activity
  GROUP BY user_id
),
months AS (
  SELECT m::date AS month
  FROM generate_series(date '2024-02-01', date '2024-05-01', interval '1 month') AS m
),
classified AS (
  -- активные в месяце
  SELECT m.month, a.user_id,
         CASE
           WHEN f.first_month = m.month THEN 'new'
           WHEN EXISTS (SELECT 1 FROM activity p
                        WHERE p.user_id = a.user_id
                          AND p.month = (m.month - interval '1 month')::date) THEN 'retained'
           ELSE 'reactivated'
         END AS kind
  FROM months m
  JOIN activity a   ON a.month = m.month
  JOIN first_seen f ON f.user_id = a.user_id
),
churned AS (
  SELECT m.month, count(*) AS churned
  FROM months m
  JOIN activity p ON p.month = (m.month - interval '1 month')::date
  WHERE NOT EXISTS (SELECT 1 FROM activity a
                    WHERE a.user_id = p.user_id AND a.month = m.month)
  GROUP BY m.month
)
SELECT m.month,
       count(c.user_id)                               AS active,
       count(*) FILTER (WHERE c.kind = 'new')         AS new,
       count(*) FILTER (WHERE c.kind = 'retained')    AS retained,
       count(*) FILTER (WHERE c.kind = 'reactivated') AS reactivated,
       coalesce(ch.churned, 0)                        AS churned
FROM months m
LEFT JOIN classified c ON c.month = m.month
LEFT JOIN churned ch   ON ch.month = m.month
GROUP BY m.month, ch.churned
ORDER BY m.month;
