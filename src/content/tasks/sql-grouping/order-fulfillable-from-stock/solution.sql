WITH on_hand AS (
  SELECT product_id, sum(qty) AS qty
  FROM stock
  GROUP BY product_id
),
need AS (
  SELECT i.order_id, i.product_id, sum(i.qty) AS qty
  FROM order_items i
  JOIN orders o ON o.id = i.order_id
  WHERE o.status = 'new'
  GROUP BY i.order_id, i.product_id
)
SELECT n.order_id,
       count(*) AS items,
       count(*) FILTER (WHERE coalesce(h.qty, 0) < n.qty) AS missing,
       bool_and(coalesce(h.qty, 0) >= n.qty)              AS can_ship
FROM need n
LEFT JOIN on_hand h ON h.product_id = n.product_id
GROUP BY n.order_id
ORDER BY n.order_id;
