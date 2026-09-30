WITH lines AS (
  SELECT DISTINCT i.order_id, i.product_id
  FROM order_items i
  JOIN orders o ON o.id = i.order_id
  WHERE o.status = 'paid'
    AND i.product_id IS NOT NULL
)
SELECT pa.name AS product_a,
       pb.name AS product_b,
       count(*) AS orders_together
FROM lines a
JOIN lines b    ON b.order_id = a.order_id AND a.product_id < b.product_id
JOIN products pa ON pa.id = a.product_id
JOIN products pb ON pb.id = b.product_id
GROUP BY a.product_id, b.product_id, pa.name, pb.name
HAVING count(*) >= 2
ORDER BY orders_together DESC, a.product_id, b.product_id;
