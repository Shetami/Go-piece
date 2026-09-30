-- План-факт по узлам с планом, апрель—июнь 2024, с нарастающим итогом.
-- Колонки: region, month, plan, fact, plan_ytd, fact_ytd, pct_ytd. Порядок: region, month.
SELECT r.name AS region,
       p.month,
       p.amount AS plan,
       sum(s.amount) AS fact,
       p.amount AS plan_ytd,
       sum(s.amount) AS fact_ytd,
       round(100 * sum(s.amount) / p.amount, 1) AS pct_ytd
FROM plans p
JOIN regions r ON r.id = p.region_id
JOIN stores sh ON sh.region_id IN (SELECT id FROM regions WHERE parent_id = r.id)
JOIN sales s   ON s.store_id = sh.id AND date_trunc('month', s.sold_at) = p.month
GROUP BY r.name, p.month, p.amount
ORDER BY region, p.month;
