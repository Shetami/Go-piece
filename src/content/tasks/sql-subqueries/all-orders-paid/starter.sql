-- Клиенты, у которых есть хотя бы один заказ и ВСЕ заказы оплачены (status = 'paid').
-- Колонка: name. Порядок: по name.
SELECT DISTINCT c.name
FROM customers c
JOIN orders o ON o.customer_id = c.id
WHERE o.status = 'paid'
ORDER BY c.name;
