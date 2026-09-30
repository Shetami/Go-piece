WITH RECURSIVE calendar AS (
  SELECT date '2024-02-26' AS day
  UNION ALL
  SELECT (day + 1)::date
  FROM calendar
  WHERE day < date '2024-03-03'
),
daily AS (
  SELECT created_at::date AS day, count(*) AS orders, sum(amount) AS revenue
  FROM orders
  WHERE status = 'paid'
    AND created_at >= date '2024-02-26'
    AND created_at <  date '2024-03-04'
  GROUP BY created_at::date
)
SELECT c.day,
       coalesce(d.orders, 0)  AS orders,
       coalesce(d.revenue, 0) AS revenue,
       sum(coalesce(d.revenue, 0)) OVER (ORDER BY c.day) AS running
FROM calendar c
LEFT JOIN daily d ON d.day = c.day
ORDER BY c.day;
