SELECT p.id, p.name
FROM products p
WHERE NOT EXISTS (
  SELECT 1 FROM order_items i WHERE i.product_id = p.id
)
ORDER BY p.id;
