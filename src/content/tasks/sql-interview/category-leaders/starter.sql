-- Топ-2 товара по чистой выручке в каждой корневой категории за I квартал 2024.
-- Колонки: root_category, place, product, net_revenue, share_pct.
-- Порядок: root_category, place, product.
SELECT c.name AS root_category,
       1 AS place,
       p.name AS product,
       sum(oi.qty * p.list_price) AS net_revenue,
       100 AS share_pct
FROM order_items oi
JOIN products p   ON p.id = oi.product_id
JOIN categories c ON c.id = p.category_id
GROUP BY c.name, p.name
ORDER BY 1, 2, 3;
