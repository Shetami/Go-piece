SELECT c.id, c.name
FROM customers c
WHERE EXISTS (
  SELECT 1
  FROM orders o
  JOIN order_items i ON i.order_id = o.id
  JOIN products p    ON p.id = i.product_id
  JOIN categories k  ON k.id = p.category_id
  WHERE o.customer_id = c.id
    AND o.status <> 'cancelled'
    AND k.name = 'Кофе'
)
ORDER BY c.id;
