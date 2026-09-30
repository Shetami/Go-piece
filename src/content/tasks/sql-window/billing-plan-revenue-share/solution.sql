WITH by_plan AS (
  SELECT to_char(paid_at, 'YYYY-MM') AS month, plan,
         sum(CASE status WHEN 'paid' THEN amount ELSE -amount END) AS revenue
  FROM payments
  WHERE status IN ('paid', 'refunded')
  GROUP BY 1, 2
)
SELECT month, plan, revenue,
       round(100.0 * revenue / sum(revenue) OVER (PARTITION BY month), 1) AS share_pct
FROM by_plan
ORDER BY month, revenue DESC, plan;
