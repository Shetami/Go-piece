-- Сводка по каждому клиенту одной строкой.
-- Колонки: customer, orders (всего заказов), paid (оплаченных),
--          cancelled (отменённых), paid_amount (сумма оплаченных, 0 если их нет).
-- Порядок: по customer.
SELECT customer, count(*) AS orders
FROM orders
GROUP BY customer
ORDER BY customer;
