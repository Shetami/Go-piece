SELECT c.name AS category, t.rank, t.product, t.units
FROM categories c
LEFT JOIN LATERAL (
  SELECT row_number() OVER (ORDER BY sum(i.qty) DESC, p.id) AS rank,
         p.name     AS product,
         sum(i.qty) AS units
  FROM products p
  JOIN order_items i ON i.product_id = p.id
  JOIN orders o      ON o.id = i.order_id
  WHERE p.category_id = c.id
    AND o.status = 'delivered'
    AND o.created_at >= '2024-05-01' AND o.created_at < '2024-06-01'
  GROUP BY p.id, p.name
  ORDER BY units DESC, p.id
  LIMIT 2
) t ON true
ORDER BY c.name, t.rank;
