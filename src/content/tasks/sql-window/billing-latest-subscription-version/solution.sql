WITH latest AS (
  SELECT customer_id, plan_code, status,
         row_number() OVER (
           PARTITION BY customer_id
           ORDER BY updated_at DESC, id DESC
         ) AS rn
  FROM subscription_versions
),
current_active AS (
  SELECT customer_id, plan_code
  FROM latest
  WHERE rn = 1 AND status = 'active'
)
SELECT p.name,
       count(c.customer_id)                  AS customers,
       round(coalesce(sum(p.monthly_price) FILTER (WHERE c.customer_id IS NOT NULL), 0), 2) AS mrr
FROM plans p
LEFT JOIN current_active c ON c.plan_code = p.code
GROUP BY p.code, p.name
ORDER BY mrr DESC, p.name;
