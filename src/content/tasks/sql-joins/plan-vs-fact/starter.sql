-- План-факт за первый квартал 2024 года по регионам и месяцам.
-- Колонки: region, month ('YYYY-MM'), plan, fact, diff (fact - plan),
--          pct (fact / plan * 100, round(…, 1); NULL, если плана нет или он 0).
-- Порядок: по region, затем по month.
SELECT region,
       to_char(month, 'YYYY-MM') AS month,
       p.amount AS plan,
       f.amount AS fact,
       f.amount - p.amount AS diff,
       round(100 * f.amount / p.amount, 1) AS pct
FROM sales_plan p
NATURAL FULL JOIN sales_fact f
WHERE p.month >= '2024-01-01' AND p.month < '2024-04-01'
ORDER BY region, month;
