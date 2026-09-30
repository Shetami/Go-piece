-- Клиенты, купившие хотя бы один товар из категории «Кофе» (отменённые заказы не в счёт).
-- Колонки: id, name. Порядок: по id. Каждый клиент — одной строкой.
SELECT c.id, c.name
FROM customers c
JOIN orders o      ON o.customer_id = c.id
JOIN order_items i ON i.order_id = o.id
JOIN products p    ON p.id = i.product_id
JOIN categories k  ON k.id = p.category_id
WHERE k.name = 'Кофе'
ORDER BY c.id;
