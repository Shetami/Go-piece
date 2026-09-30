-- Доля клиентов, вернувшихся за повторной покупкой в течение 30 дней после первой.
-- Колонки: cohort ('YYYY-MM' первого заказа), customers, repeaters, repeat_pct (round(…, 1)).
-- Порядок: по cohort.
SELECT to_char(a.created_at, 'YYYY-MM') AS cohort,
       count(DISTINCT a.customer_id) AS customers,
       count(DISTINCT b.customer_id) AS repeaters,
       round(100.0 * count(DISTINCT b.customer_id) / count(DISTINCT a.customer_id), 1) AS repeat_pct
FROM orders a
LEFT JOIN orders b ON b.customer_id = a.customer_id AND b.id <> a.id
                  AND b.created_at < a.created_at + interval '30 days'
GROUP BY 1
ORDER BY 1;
