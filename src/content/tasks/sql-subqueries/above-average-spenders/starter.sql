-- Клиенты, чья сумма неотменённых заказов больше средней такой суммы по клиентам.
-- Колонки: name, total. Порядок: total DESC, name.
SELECT c.name, sum(o.amount) AS total
FROM customers c
JOIN orders o ON o.customer_id = c.id
GROUP BY c.name
HAVING sum(o.amount) > (SELECT avg(amount) * 2 FROM orders)
ORDER BY total DESC, c.name;
