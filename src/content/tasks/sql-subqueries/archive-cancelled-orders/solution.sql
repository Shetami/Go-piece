WITH moved AS (
  DELETE FROM orders
  WHERE status = 'cancelled'
    AND created_at < timestamp '2024-01-01'
  RETURNING *
),
archived AS (
  INSERT INTO orders_archive
  SELECT id, customer_id, amount, status, created_at
  FROM moved
  RETURNING customer_id, amount
),
per_customer AS (
  SELECT customer_id, count(*) AS moved, sum(amount) AS moved_amount
  FROM archived
  GROUP BY customer_id
)
SELECT c.name,
       m.moved,
       m.moved_amount,
       -- основной запрос видит таблицы в состоянии ДО удаления и вставки
       (SELECT count(*) FROM orders o         WHERE o.customer_id = c.id) - m.moved AS left_in_orders,
       (SELECT count(*) FROM orders_archive a WHERE a.customer_id = c.id) + m.moved AS in_archive
FROM per_customer m
JOIN customers c ON c.id = m.customer_id
ORDER BY c.name;
