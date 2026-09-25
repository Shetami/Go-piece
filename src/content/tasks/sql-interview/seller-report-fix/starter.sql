-- Отчёт по всем продавцам: доставленные заказы, выручка, средний рейтинг товаров.
-- Колонки: id, seller, orders_count, revenue, avg_rating.
-- Порядок: revenue по убыванию, затем id.
SELECT s.id,
       s.name AS seller,
       count(o.id) AS orders_count,
       sum(oi.qty * oi.price) AS revenue,
       round(avg(r.rating), 2) AS avg_rating
FROM sellers s
LEFT JOIN products p     ON p.seller_id = s.id
LEFT JOIN order_items oi ON oi.product_id = p.id
LEFT JOIN orders o       ON o.id = oi.order_id
LEFT JOIN reviews r      ON r.product_id = p.id
WHERE o.status = 'delivered'
GROUP BY s.id, s.name
ORDER BY revenue DESC, s.id;
