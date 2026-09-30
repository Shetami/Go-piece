WITH expected AS (
  SELECT i.order_id, i.product_id, i.qty, i.unit_price,
         coalesce(
           (SELECT min(pr.promo_price)
            FROM promos pr
            WHERE pr.product_id = i.product_id
              AND o.created_at >= pr.starts_at
              AND o.created_at <  pr.ends_at),
           (SELECT ph.price
            FROM price_history ph
            WHERE ph.product_id = i.product_id
              AND o.created_at >= ph.valid_from
              AND (ph.valid_to IS NULL OR o.created_at < ph.valid_to))
         ) AS expected_price
  FROM order_items i
  JOIN orders o ON o.id = i.order_id
  WHERE o.status = 'paid'
)
SELECT e.order_id,
       p.name AS product,
       e.qty,
       e.unit_price AS charged,
       e.expected_price AS expected,
       (e.unit_price - e.expected_price) * e.qty AS diff_total
FROM expected e
JOIN products p ON p.id = e.product_id
WHERE e.unit_price IS DISTINCT FROM e.expected_price
ORDER BY e.order_id, product;
