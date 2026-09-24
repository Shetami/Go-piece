-- Клиенты, у которых больше двух заказов, и сумма их заказов
SELECT customer, count(*) AS orders, sum(amount) AS total
FROM orders
WHERE count(*) > 2
GROUP BY customer
ORDER BY customer;
