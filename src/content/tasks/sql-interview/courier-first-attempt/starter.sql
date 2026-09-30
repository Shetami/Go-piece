-- Конверсия доставки с первой попытки по курьерам.
-- Колонки: courier, attempts, orders, delivered, first_try_pct. Порядок: по courier.
SELECT c.name AS courier,
       count(*) AS attempts,
       count(DISTINCT a.order_id) AS orders,
       count(*) FILTER (WHERE a.result = 'delivered') AS delivered,
       round(100.0 * count(*) FILTER (WHERE a.result = 'delivered') / count(*), 1) AS first_try_pct
FROM couriers c
JOIN attempts a ON a.courier_id = c.id
GROUP BY c.name
ORDER BY courier;
