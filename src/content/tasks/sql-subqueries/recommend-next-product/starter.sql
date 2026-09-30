-- Для каждого клиента — самый популярный (по числу разных покупателей) товар из категорий,
-- где клиент уже покупал, который он сам ещё не покупал. Ничья — по имени товара.
-- Колонки: customer, product, buyers. Порядок: customer.
SELECT DISTINCT ON (c.name) c.name AS customer, p.name AS product, count(*) AS buyers
FROM customers c
JOIN products p ON p.category IN (
       SELECT p2.category FROM order_items oi JOIN products p2 ON p2.id = oi.product_id
       WHERE oi.customer_id = c.id)
JOIN order_items x ON x.product_id = p.id
WHERE p.id NOT IN (SELECT product_id FROM order_items WHERE customer_id = c.id)
GROUP BY c.name, p.name
ORDER BY c.name, count(*) DESC, p.name;
