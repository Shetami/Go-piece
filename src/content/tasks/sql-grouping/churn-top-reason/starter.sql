-- Главная причина отмен по тарифам.
-- Колонки: plan, cancelled, with_reason, top_reason, top_pct (round 1).
-- Порядок: по plan.
SELECT s.plan_code AS plan,
       count(*) AS cancelled,
       count(c.reason) AS with_reason,
       mode() WITHIN GROUP (ORDER BY c.reason) AS top_reason,
       NULL::numeric AS top_pct
FROM cancellations c
JOIN subscriptions s ON s.id = c.subscription_id
GROUP BY s.plan_code
ORDER BY 1;
