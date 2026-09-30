-- Новые платные подписки: месяцы (январь–апрель 2024) × тарифы, с итогами.
-- Колонки: month ('YYYY-MM', последняя строка 'Итого'), basic, pro, team, total.
-- Порядок: месяцы по возрастанию, 'Итого' — последней.
SELECT to_char(created_at, 'YYYY-MM') AS month,
       count(*) FILTER (WHERE plan = 'basic') AS basic,
       count(*) FILTER (WHERE plan = 'pro')   AS pro,
       count(*) FILTER (WHERE plan = 'team')  AS team,
       count(*) AS total
FROM subscriptions
WHERE NOT is_trial
GROUP BY 1
ORDER BY 1;
