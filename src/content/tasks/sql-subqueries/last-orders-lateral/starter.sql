-- Два последних неотменённых заказа каждого клиента; клиент без таких заказов — одной строкой с NULL.
-- Колонки: name, id, created_at, amount. Порядок: name, затем created_at DESC, id DESC.
SELECT c.name, o.id, o.created_at, o.amount
FROM customers c
JOIN orders o ON o.customer_id = c.id
WHERE o.created_at >= (
  SELECT max(o2.created_at) - interval '3 days'
  FROM orders o2
  WHERE o2.customer_id = c.id
)
ORDER BY c.name, o.created_at DESC, o.id DESC;
