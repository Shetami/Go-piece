-- Все клиенты и число их заказов за 2024 год — включая тех, у кого заказов не было
SELECT c.name, count(o.id) AS orders_2024
FROM customers c
LEFT JOIN orders o
  ON o.customer_id = c.id
 AND o.created_at >= '2024-01-01' AND o.created_at < '2025-01-01'
GROUP BY c.id, c.name
ORDER BY c.name;
