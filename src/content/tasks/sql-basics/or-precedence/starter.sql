-- Крупные заказы (от 1000), которые уже оплачены или отправлены
SELECT id, customer, status, amount
FROM orders
WHERE amount >= 1000
  AND status = 'paid' OR status = 'shipped'
ORDER BY id;
