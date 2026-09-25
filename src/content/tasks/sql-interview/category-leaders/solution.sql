WITH RECURSIVE tree AS (
  SELECT id, id AS root_id
  FROM categories
  WHERE parent_id IS NULL
  UNION ALL
  SELECT c.id, t.root_id
  FROM categories c
  JOIN tree t ON c.parent_id = t.id
),
good_orders AS (
  SELECT o.id
  FROM orders o
  JOIN customers c ON c.id = o.customer_id
  WHERE NOT c.is_test
    AND o.status IN ('paid', 'shipped')
    AND o.created_at >= '2024-01-01'
    AND o.created_at <  '2024-04-01'
),
refunded AS (
  SELECT order_id, product_id, sum(qty) AS qty
  FROM refunds
  GROUP BY order_id, product_id
),
product_net AS (
  SELECT oi.product_id,
         sum((oi.qty - coalesce(r.qty, 0)) * oi.price) AS net_revenue
  FROM order_items oi
  JOIN good_orders g ON g.id = oi.order_id
  LEFT JOIN refunded r ON r.order_id = oi.order_id AND r.product_id = oi.product_id
  GROUP BY oi.product_id
),
ranked AS (
  SELECT root.name AS root_category,
         p.name    AS product,
         pn.net_revenue,
         dense_rank() OVER (PARTITION BY t.root_id ORDER BY pn.net_revenue DESC) AS place,
         round(100.0 * pn.net_revenue / sum(pn.net_revenue) OVER (PARTITION BY t.root_id), 1) AS share_pct
  FROM product_net pn
  JOIN products p      ON p.id = pn.product_id
  JOIN tree t          ON t.id = p.category_id
  JOIN categories root ON root.id = t.root_id
)
SELECT root_category, place, product, net_revenue, share_pct
FROM ranked
WHERE place <= 2
ORDER BY root_category, place, product;
