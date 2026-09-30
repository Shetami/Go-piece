WITH a AS (
  -- дубли сканирования схлопываются
  SELECT DISTINCT at.order_id, at.courier_id, at.attempted_at, at.result
  FROM attempts at
  JOIN orders o ON o.id = at.order_id
  WHERE NOT o.is_test
),
numbered AS (
  SELECT a.*,
         row_number() OVER (PARTITION BY order_id ORDER BY attempted_at) AS attempt_no
  FROM a
),
per_courier AS (
  SELECT courier_id,
         count(*)                                                    AS attempts,
         count(DISTINCT order_id)                                    AS orders,
         count(*) FILTER (WHERE result = 'delivered')                AS delivered,
         count(*) FILTER (WHERE attempt_no = 1)                      AS first_attempts,
         count(*) FILTER (WHERE attempt_no = 1 AND result = 'delivered') AS first_ok
  FROM numbered
  GROUP BY courier_id
)
SELECT c.name AS courier,
       coalesce(p.attempts, 0)  AS attempts,
       coalesce(p.orders, 0)    AS orders,
       coalesce(p.delivered, 0) AS delivered,
       round(100.0 * p.first_ok / nullif(p.first_attempts, 0), 1) AS first_try_pct
FROM couriers c
LEFT JOIN per_courier p ON p.courier_id = c.id
ORDER BY courier;
