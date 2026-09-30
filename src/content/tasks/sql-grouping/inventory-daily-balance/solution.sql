WITH opening AS (
  SELECT sku, sum(qty) AS qty
  FROM stock_moves
  WHERE moved_at < '2024-08-01'
  GROUP BY sku
),
daily AS (
  SELECT sku, moved_at::date AS day, sum(qty) AS delta
  FROM stock_moves
  WHERE moved_at >= '2024-08-01' AND moved_at < '2024-08-06'
  GROUP BY sku, moved_at::date
),
grid AS (
  SELECT p.sku, d.day::date AS day
  FROM products p
  CROSS JOIN generate_series(date '2024-08-01', date '2024-08-05', interval '1 day') AS d(day)
)
SELECT g.sku, g.day,
       coalesce(o.qty, 0)
         + sum(coalesce(dl.delta, 0)) OVER (PARTITION BY g.sku ORDER BY g.day) AS balance
FROM grid g
LEFT JOIN daily dl  ON dl.sku = g.sku AND dl.day = g.day
LEFT JOIN opening o ON o.sku = g.sku
ORDER BY g.sku, g.day;
