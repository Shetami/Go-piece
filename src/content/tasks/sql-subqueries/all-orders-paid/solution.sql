SELECT c.name
FROM customers c
WHERE EXISTS (SELECT 1 FROM orders o WHERE o.customer_id = c.id)
  AND NOT EXISTS (
    SELECT 1 FROM orders o
    WHERE o.customer_id = c.id
      AND o.status IS DISTINCT FROM 'paid'
  )
ORDER BY c.name;
