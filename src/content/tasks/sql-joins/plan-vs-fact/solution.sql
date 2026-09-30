WITH plan AS (
  SELECT region, month, amount
  FROM sales_plan
  WHERE month >= '2024-01-01' AND month < '2024-04-01'
),
fact AS (
  SELECT coalesce(region, 'Без региона') AS region, month, sum(amount) AS amount
  FROM sales_fact
  WHERE month >= '2024-01-01' AND month < '2024-04-01'
  GROUP BY 1, 2
)
SELECT region,
       to_char(month, 'YYYY-MM')                        AS month,
       coalesce(p.amount, 0)                             AS plan,
       coalesce(f.amount, 0)                             AS fact,
       coalesce(f.amount, 0) - coalesce(p.amount, 0)     AS diff,
       round(100 * coalesce(f.amount, 0) / nullif(p.amount, 0), 1) AS pct
FROM plan p
FULL JOIN fact f USING (region, month)
ORDER BY region, month;
