WITH popularity AS (
  SELECT p.id, p.name, p.category,
         count(DISTINCT oi.customer_id) AS buyers
  FROM products p
  LEFT JOIN order_items oi ON oi.product_id = p.id
  GROUP BY p.id, p.name, p.category
),
candidates AS (
  SELECT c.name AS customer, pop.name AS product, pop.buyers,
         row_number() OVER (PARTITION BY c.id ORDER BY pop.buyers DESC, pop.name) AS rn
  FROM customers c
  JOIN popularity pop
    ON EXISTS (                                   -- покупал что-то из этой категории
         SELECT 1 FROM order_items oi
         JOIN products p ON p.id = oi.product_id
         WHERE oi.customer_id = c.id AND p.category = pop.category)
   AND NOT EXISTS (                               -- но не этот товар
         SELECT 1 FROM order_items oi
         WHERE oi.customer_id = c.id AND oi.product_id = pop.id)
)
SELECT customer, product, buyers
FROM candidates
WHERE rn = 1
ORDER BY customer;
