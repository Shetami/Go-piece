-- Заказы клиентов из Казани
SELECT id, amount
FROM orders
WHERE customer_id IN (SELECT id FROM customers WHERE city = 'Казань')
ORDER BY id;
