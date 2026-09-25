WITH months AS (
  SELECT m::date AS month
  FROM generate_series(date '2024-01-01', date '2024-05-01', interval '1 month') AS m
),
mrr AS (
  SELECT c.id AS customer_id, m.month,
         coalesce(sum(p.monthly_price), 0) AS mrr
  FROM customers c
  CROSS JOIN months m
  LEFT JOIN subscriptions s ON s.customer_id = c.id
                           AND s.started_on <= m.month
                           AND (s.ended_on IS NULL OR s.ended_on > m.month)
  LEFT JOIN plans p ON p.id = s.plan_id
  GROUP BY c.id, m.month
),
moves AS (
  SELECT month, mrr,
         lag(mrr) OVER (PARTITION BY customer_id ORDER BY month) AS prev
  FROM mrr
)
SELECT month,
       sum(CASE WHEN prev = 0 AND mrr > 0      THEN mrr        ELSE 0 END) AS new_mrr,
       sum(CASE WHEN prev > 0 AND mrr > prev   THEN mrr - prev ELSE 0 END) AS expansion_mrr,
       sum(CASE WHEN mrr > 0  AND mrr < prev   THEN prev - mrr ELSE 0 END) AS contraction_mrr,
       sum(CASE WHEN prev > 0 AND mrr = 0      THEN prev       ELSE 0 END) AS churned_mrr,
       sum(mrr) AS total_mrr
FROM moves
WHERE month >= date '2024-02-01'
GROUP BY month
ORDER BY month;
