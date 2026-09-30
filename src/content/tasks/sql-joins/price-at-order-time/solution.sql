WITH prices AS (
  SELECT DISTINCT product_id, price, valid_from, valid_to
  FROM price_history
)
SELECT o.id AS order_id,
       coalesce(sum(i.qty * p.price), 0)             AS revenue,
       count(*) FILTER (WHERE p.product_id IS NULL)  AS unpriced_items
FROM orders o
JOIN order_items i ON i.order_id = o.id
LEFT JOIN prices p
  ON p.product_id = i.product_id
 AND p.valid_from <= o.created_at
 AND (p.valid_to IS NULL OR o.created_at < p.valid_to)
GROUP BY o.id
ORDER BY o.id;
