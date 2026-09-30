WITH days AS (
  SELECT d::date AS day
  FROM generate_series(date '2024-03-01', date '2024-03-05', interval '1 day') AS d
),
daily AS (
  SELECT store_id, sold_at::date AS day, count(*) AS checks, sum(amount) AS revenue
  FROM sales
  WHERE sold_at >= '2024-03-01' AND sold_at < '2024-03-06'
  GROUP BY store_id, sold_at::date
)
SELECT d.day, s.name AS store,
       coalesce(x.checks, 0)  AS checks,
       coalesce(x.revenue, 0) AS revenue
FROM days d
CROSS JOIN stores s
LEFT JOIN daily x ON x.store_id = s.id AND x.day = d.day
WHERE s.opened_on <= d.day
  AND (s.closed_on IS NULL OR d.day < s.closed_on)
ORDER BY d.day, s.name;
