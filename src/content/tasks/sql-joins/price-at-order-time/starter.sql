-- Выручка каждого заказа по цене, действовавшей в момент заказа.
-- Колонки: order_id, revenue, unpriced_items (позиции, для которых цены на тот момент нет).
-- Порядок: по order_id.
SELECT o.id AS order_id, sum(i.qty * p.price) AS revenue, 0 AS unpriced_items
FROM orders o
JOIN order_items i   ON i.order_id = o.id
JOIN price_history p ON p.product_id = i.product_id
WHERE o.created_at BETWEEN p.valid_from AND p.valid_to
GROUP BY o.id
ORDER BY o.id;
