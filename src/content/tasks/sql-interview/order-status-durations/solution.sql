WITH marked AS (
  SELECT l.*,
         lag(status) OVER (PARTITION BY order_id ORDER BY changed_at) AS prev_status
  FROM status_log l
),
visits AS (
  -- повтор того же статуса подряд — не новый визит
  SELECT order_id, status, changed_at,
         lead(changed_at) OVER (PARTITION BY order_id ORDER BY changed_at) AS left_at
  FROM marked
  WHERE prev_status IS DISTINCT FROM status
),
per_order AS (
  SELECT v.order_id, v.status,
         sum(extract(epoch FROM coalesce(v.left_at, timestamp '2024-10-10 12:00') - v.changed_at) / 3600) AS hours
  FROM visits v
  GROUP BY v.order_id, v.status
)
SELECT s.code AS status,
       count(p.order_id)       AS orders,
       round(avg(p.hours), 1)  AS avg_hours,
       round(max(p.hours), 1)  AS max_hours
FROM statuses s
LEFT JOIN per_order p ON p.status = s.code
WHERE NOT s.is_final
GROUP BY s.code, s.sort_order
ORDER BY s.sort_order;
