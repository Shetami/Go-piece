-- Выручка по ISO-неделям без отменённых заказов.
-- Колонки: week ('2025-W01'), week_start (date, понедельник), orders, revenue.
-- Порядок: по week_start.
SELECT to_char(created_at, 'YYYY-"W"IW') AS week,
       date_trunc('week', created_at)::date AS week_start,
       count(*) AS orders,
       sum(amount) AS revenue
FROM orders
WHERE status <> 'cancelled'
GROUP BY 1, 2
ORDER BY week_start;
