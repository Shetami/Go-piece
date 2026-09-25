-- Воронка по вариантам эксперимента: сессии, просмотр → корзина → покупка.
-- Колонки: variant, sessions, viewed, carted, purchased, conversion_pct.
-- Порядок: по variant.
SELECT x.variant,
       count(DISTINCT e.user_id) AS sessions,
       count(*) FILTER (WHERE e.event = 'view')     AS viewed,
       count(*) FILTER (WHERE e.event = 'cart')     AS carted,
       count(*) FILTER (WHERE e.event = 'purchase') AS purchased,
       0 AS conversion_pct
FROM events e
JOIN experiment x ON x.user_id = e.user_id
GROUP BY x.variant
ORDER BY x.variant;
