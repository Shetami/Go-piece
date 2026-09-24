-- Клиенты, у которых больше двух заказов, и сумма их заказов
SELECT customer, count(*) AS orders, sum(amount) AS total
FROM orders
GROUP BY customer
HAVING count(*) > 2
ORDER BY customer;
