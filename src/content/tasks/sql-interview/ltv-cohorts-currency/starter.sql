-- LTV за первые 30 дней по месячным когортам регистрации, в долларах.
-- Колонки: cohort, users, paying_users, revenue_usd, ltv30_usd. Порядок: по cohort.
SELECT date_trunc('month', u.signup_on)::date AS cohort,
       count(DISTINCT u.id) AS users,
       count(DISTINCT o.user_id) AS paying_users,
       round(sum(o.amount * fx.rate_to_usd), 2) AS revenue_usd,
       round(avg(o.amount * fx.rate_to_usd), 2) AS ltv30_usd
FROM users u
JOIN orders o     ON o.user_id = u.id AND o.created_on <= u.signup_on + 30
JOIN fx_monthly fx ON fx.currency = o.currency
                  AND fx.month = date_trunc('month', u.signup_on)
GROUP BY 1
ORDER BY cohort;
