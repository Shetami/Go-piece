-- Сколько часов заказы проводят в каждом нефинальном статусе.
-- Колонки: status, orders, avg_hours, max_hours. Порядок: по statuses.sort_order.
WITH d AS (
  SELECT order_id, status,
         lead(changed_at) OVER (PARTITION BY order_id ORDER BY changed_at) - changed_at AS spent
  FROM status_log
)
SELECT d.status,
       count(*) AS orders,
       round(avg(extract(epoch FROM d.spent) / 3600), 1) AS avg_hours,
       round(max(extract(epoch FROM d.spent) / 3600), 1) AS max_hours
FROM d
JOIN statuses s ON s.code = d.status
WHERE NOT s.is_final
GROUP BY d.status, s.sort_order
ORDER BY s.sort_order;
