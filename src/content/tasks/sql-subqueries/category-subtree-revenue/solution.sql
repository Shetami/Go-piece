WITH RECURSIVE subtree AS (
  -- каждая категория — предок самой себя
  SELECT id AS root_id, id
  FROM categories

  UNION ALL

  SELECT s.root_id, c.id
  FROM subtree s
  JOIN categories c ON c.parent_id = s.id
),
product_revenue AS (
  SELECT p.id, p.category_id, coalesce(sum(s.qty * s.price), 0) AS revenue
  FROM products p
  LEFT JOIN sales s ON s.product_id = p.id
  GROUP BY p.id, p.category_id
)
SELECT c.name,
       coalesce(sum(pr.revenue), 0) AS revenue,
       count(pr.id)                 AS products
FROM categories c
JOIN subtree t             ON t.root_id = c.id
LEFT JOIN product_revenue pr ON pr.category_id = t.id
GROUP BY c.id, c.name
ORDER BY revenue DESC, c.name;
