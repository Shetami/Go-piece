-- Для каждого клиента: на сколько заказал и сколько оплатил
SELECT c.name,
       coalesce(o.total, 0) AS ordered,
       coalesce(p.total, 0) AS paid
FROM customers c
LEFT JOIN (
  SELECT customer_id, sum(amount) AS total FROM orders GROUP BY customer_id
) o ON o.customer_id = c.id
LEFT JOIN (
  SELECT customer_id, sum(amount) AS total FROM payments GROUP BY customer_id
) p ON p.customer_id = c.id
ORDER BY c.name;
