-- Клиенты, купившие хотя бы по товару из каждой активной категории.
-- Колонки: customer, orders (оплаченных заказов). Порядок: по customer.
SELECT o.customer, count(*) AS orders
FROM orders o
JOIN order_items i ON i.order_id = o.id
JOIN products p ON p.id = i.product_id
GROUP BY o.customer
HAVING count(DISTINCT p.category_id) = (SELECT count(*) FROM categories WHERE is_active)
ORDER BY o.customer;
