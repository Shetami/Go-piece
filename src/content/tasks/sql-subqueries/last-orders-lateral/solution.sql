SELECT c.name, o.id, o.created_at, o.amount
FROM customers c
LEFT JOIN LATERAL (
  SELECT id, created_at, amount
  FROM orders
  WHERE customer_id = c.id
    AND status <> 'cancelled'
  ORDER BY created_at DESC, id DESC
  LIMIT 2
) o ON true
ORDER BY c.name, o.created_at DESC, o.id DESC;
