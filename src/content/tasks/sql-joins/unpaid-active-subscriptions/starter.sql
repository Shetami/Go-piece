-- Подписки, действовавшие в июне 2024 хотя бы один день, за июнь которых нет оплаченного счёта.
-- Колонки: customer, subscription_id, plan. Порядок: по subscription_id.
SELECT c.name AS customer, s.id AS subscription_id, s.plan
FROM subscriptions s
JOIN customers c ON c.id = s.customer_id
WHERE s.ended_on IS NULL
  AND s.id NOT IN (SELECT subscription_id FROM invoices WHERE period_start = '2024-06-01')
ORDER BY s.id;
