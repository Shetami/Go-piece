-- Для каждого клиента: на сколько заказал и сколько оплатил
SELECT c.name, sum(o.amount) AS ordered, sum(p.amount) AS paid
FROM customers c
JOIN orders o ON o.customer_id = c.id
JOIN payments p ON p.customer_id = c.id
GROUP BY c.id, c.name
ORDER BY c.name;
