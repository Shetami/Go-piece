-- Текущий MRR по тарифам: у каждого клиента действует последняя версия подписки
-- (по updated_at, при равенстве — больший id); отменённые не считаются.
-- Все тарифы из справочника, даже без клиентов.
-- Колонки: name, customers, mrr (round 2). Порядок: mrr по убыванию, затем name.
SELECT p.name, count(*) AS customers, round(sum(p.monthly_price), 2) AS mrr
FROM subscription_versions v
JOIN plans p ON p.code = v.plan_code
WHERE v.status = 'active'
GROUP BY p.name
ORDER BY mrr DESC, p.name;
