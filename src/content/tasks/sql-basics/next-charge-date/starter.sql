-- Ближайшее списание по каждой подписке на 2024-03-30 (сегодня).
-- Колонки: id, customer, next_charge (date, NULL — списаний больше не будет).
-- Порядок: next_charge по возрастанию, NULL в конце, при равенстве — id.
SELECT id, customer, (started_on + interval '1 month')::date AS next_charge
FROM subscriptions
ORDER BY next_charge, id;
