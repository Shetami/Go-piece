WITH categories AS (
  SELECT DISTINCT category
  FROM products
  WHERE is_active
)
SELECT k.category, c.name
FROM categories k
CROSS JOIN customers c
WHERE NOT EXISTS (
  SELECT 1
  FROM products p
  WHERE p.category = k.category
    AND p.is_active
    AND NOT EXISTS (
      SELECT 1
      FROM purchases x
      WHERE x.customer_id = c.id
        AND x.product_id = p.id
        AND NOT x.returned
    )
)
ORDER BY k.category, c.name;
