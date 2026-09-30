-- Одним запросом: перенести отменённые заказы, созданные до 2024-01-01, из orders в orders_archive
-- и вернуть по каждому клиенту, у которого что-то перенесли:
-- name, moved, moved_amount, left_in_orders (осталось в orders), in_archive (всего в архиве). Порядок: name.
WITH moved AS (
  DELETE FROM orders
  WHERE status = 'cancelled'
    AND created_at < timestamp '2024-01-01'
  RETURNING *
),
archived AS (
  INSERT INTO orders_archive
  SELECT id, customer_id, amount, status, created_at FROM moved
)
SELECT c.name,
       count(*)      AS moved,
       sum(m.amount) AS moved_amount,
       (SELECT count(*) FROM orders o         WHERE o.customer_id = c.id) AS left_in_orders,
       (SELECT count(*) FROM orders_archive a WHERE a.customer_id = c.id) AS in_archive
FROM moved m
JOIN customers c ON c.id = m.customer_id
GROUP BY c.id, c.name
ORDER BY c.name;
