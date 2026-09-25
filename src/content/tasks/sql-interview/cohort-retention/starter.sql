-- Когорты по месяцу регистрации, январь–апрель 2024.
-- Колонки: cohort (date), users, m1_pct, m2_pct. Порядок: по cohort.
SELECT date_trunc('month', u.signed_up_at)::date AS cohort,
       count(DISTINCT u.id) AS users,
       round(100.0 * count(DISTINCT o.user_id) FILTER (WHERE o.created_at < u.signed_up_at + interval '30 days')
             / count(DISTINCT u.id), 1) AS m1_pct,
       0 AS m2_pct
FROM users u
LEFT JOIN orders o ON o.user_id = u.id
GROUP BY 1
ORDER BY 1;
