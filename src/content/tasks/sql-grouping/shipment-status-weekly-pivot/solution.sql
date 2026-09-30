WITH norm AS (
  SELECT id,
         date_trunc('week', created_at::timestamp)::date AS week,
         lower(trim(status)) AS st
  FROM shipments
)
SELECT w.week::date AS week,
       count(n.id) AS total,
       count(n.id) FILTER (WHERE n.st = 'delivered')  AS delivered,
       count(n.id) FILTER (WHERE n.st = 'in_transit') AS in_transit,
       count(n.id) FILTER (WHERE n.st = 'lost')       AS lost,
       count(n.id) FILTER (WHERE n.st IS NULL
                              OR n.st NOT IN ('delivered', 'in_transit', 'lost')) AS other
FROM generate_series(date '2024-04-01', date '2024-04-29', interval '1 week') AS w(week)
LEFT JOIN norm n ON n.week = w.week::date
GROUP BY w.week
ORDER BY w.week;
