WITH totals AS (
  SELECT customer_id, sum(amount) AS total
  FROM orders
  WHERE status <> 'cancelled'
  GROUP BY customer_id
)
SELECT c.name, t.total
FROM totals t
JOIN customers c ON c.id = t.customer_id
WHERE t.total > (SELECT avg(total) FROM totals)
ORDER BY t.total DESC, c.name;
