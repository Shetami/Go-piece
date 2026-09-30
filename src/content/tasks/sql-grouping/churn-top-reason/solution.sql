WITH last_cancel AS (
  SELECT DISTINCT ON (subscription_id) subscription_id, reason
  FROM cancellations
  ORDER BY subscription_id, cancelled_at DESC, id DESC
),
per_sub AS (
  SELECT s.plan_code, l.reason
  FROM last_cancel l
  JOIN subscriptions s ON s.id = l.subscription_id
),
totals AS (
  SELECT plan_code, count(*) AS cancelled, count(reason) AS with_reason
  FROM per_sub
  GROUP BY plan_code
),
reasons AS (
  SELECT plan_code, reason, count(*) AS cnt,
         row_number() OVER (PARTITION BY plan_code ORDER BY count(*) DESC, reason) AS rn
  FROM per_sub
  WHERE reason IS NOT NULL
  GROUP BY plan_code, reason
)
SELECT p.code AS plan,
       coalesce(t.cancelled, 0)   AS cancelled,
       coalesce(t.with_reason, 0) AS with_reason,
       r.reason                   AS top_reason,
       round(100.0 * r.cnt / t.with_reason, 1) AS top_pct
FROM plans p
LEFT JOIN totals t  ON t.plan_code = p.code
LEFT JOIN reasons r ON r.plan_code = p.code AND r.rn = 1
ORDER BY p.code;
