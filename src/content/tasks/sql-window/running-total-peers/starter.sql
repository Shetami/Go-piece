-- Нарастающий итог выручки: для каждого заказа — сумма его и всех предыдущих.
-- Заказы идут в порядке id.
SELECT id, created_on, amount,
       sum(amount) OVER (ORDER BY created_on) AS running_total
FROM orders
ORDER BY id;
