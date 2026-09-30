WITH rec AS (
  SELECT product_id, sum(qty) AS qty FROM receipts GROUP BY product_id
),
sold AS (
  SELECT i.product_id, sum(i.qty) AS qty
  FROM order_items i
  JOIN orders o ON o.id = i.order_id
  WHERE o.status = 'paid'
  GROUP BY i.product_id
),
ret AS (
  SELECT product_id, sum(qty) AS qty FROM returns GROUP BY product_id
)
SELECT p.sku,
       coalesce(rec.qty, 0)  AS received,
       coalesce(sold.qty, 0) AS sold,
       coalesce(ret.qty, 0)  AS returned,
       coalesce(rec.qty, 0) - coalesce(sold.qty, 0) + coalesce(ret.qty, 0) AS on_hand
FROM products p
LEFT JOIN rec  ON rec.product_id  = p.id
LEFT JOIN sold ON sold.product_id = p.id
LEFT JOIN ret  ON ret.product_id  = p.id
ORDER BY p.sku;
