-- Счёт за февраль 2024 пропорционально активным дням.
-- Колонки: customer, plan, days, amount (round(…, 2)).
-- Порядок: customer, затем plan.
SELECT s.customer, p.name AS plan,
       coalesce(s.ended_on, '2024-02-29') - s.started_on AS days,
       round(p.monthly_price * (coalesce(s.ended_on, '2024-02-29') - s.started_on) / 30, 2) AS amount
FROM subscriptions s
JOIN plans p ON p.id = s.plan_id
WHERE s.started_on BETWEEN '2024-02-01' AND '2024-02-29'
ORDER BY s.customer, plan;
