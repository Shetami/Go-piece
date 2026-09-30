-- Склейка дублей клиентов по email или телефону.
-- Колонки: master_id, master_name, accounts, orders, revenue. Порядок: по master_id.
SELECT min(c.id) AS master_id,
       min(c.name) AS master_name,
       count(DISTINCT c.id) AS accounts,
       count(o.id) AS orders,
       sum(o.amount) AS revenue
FROM customers c
LEFT JOIN orders o ON o.customer_id = c.id
GROUP BY c.email
HAVING count(DISTINCT c.id) > 1
ORDER BY master_id;
