-- Отчёт по всем продавцам: доставленные заказы, выручка, средний рейтинг товаров.
-- Колонки: id, seller, orders_count, revenue, avg_rating.
-- Порядок: revenue по убыванию, затем id.
WITH sales AS (
  SELECT p.seller_id,
         count(DISTINCT o.id)   AS orders_count,
         sum(oi.qty * oi.price) AS revenue
  FROM products p
  JOIN order_items oi ON oi.product_id = p.id
  JOIN orders o       ON o.id = oi.order_id
  WHERE o.status = 'delivered'
  GROUP BY p.seller_id
),
ratings AS (
  SELECT p.seller_id, round(avg(r.rating), 2) AS avg_rating
  FROM products p
  JOIN reviews r ON r.product_id = p.id
  GROUP BY p.seller_id
)
SELECT s.id,
       s.name AS seller,
       coalesce(sa.orders_count, 0) AS orders_count,
       coalesce(sa.revenue, 0)      AS revenue,
       ra.avg_rating
FROM sellers s
LEFT JOIN sales sa   ON sa.seller_id = s.id
LEFT JOIN ratings ra ON ra.seller_id = s.id
ORDER BY revenue DESC, s.id;
