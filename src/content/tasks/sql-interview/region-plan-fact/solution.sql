WITH RECURSIVE subtree AS (
  SELECT id AS root_id, id AS node_id FROM regions
  UNION ALL
  SELECT s.root_id, r.id
  FROM subtree s
  JOIN regions r ON r.parent_id = s.node_id
),
months AS (
  SELECT m::date AS month
  FROM generate_series(date '2024-04-01', date '2024-06-01', interval '1 month') AS m
),
grid AS (
  SELECT DISTINCT p.region_id, m.month
  FROM plans p
  CROSS JOIN months m
),
fact AS (
  SELECT st.root_id AS region_id,
         date_trunc('month', s.sold_at)::date AS month,
         sum(s.amount) AS amount
  FROM sales s
  JOIN stores sh  ON sh.id = s.store_id
  JOIN subtree st ON st.node_id = sh.region_id
  GROUP BY st.root_id, date_trunc('month', s.sold_at)
),
monthly AS (
  SELECT g.region_id, g.month,
         coalesce(p.amount, 0) AS plan,
         coalesce(f.amount, 0) AS fact
  FROM grid g
  LEFT JOIN plans p ON p.region_id = g.region_id AND p.month = g.month
  LEFT JOIN fact f  ON f.region_id = g.region_id AND f.month = g.month
),
running AS (
  SELECT m.*,
         sum(plan) OVER w AS plan_ytd,
         sum(fact) OVER w AS fact_ytd
  FROM monthly m
  WINDOW w AS (PARTITION BY region_id ORDER BY month)
)
SELECT r.name AS region, x.month, x.plan, x.fact, x.plan_ytd, x.fact_ytd,
       round(100 * x.fact_ytd / nullif(x.plan_ytd, 0), 1) AS pct_ytd
FROM running x
JOIN regions r ON r.id = x.region_id
ORDER BY region, x.month;
