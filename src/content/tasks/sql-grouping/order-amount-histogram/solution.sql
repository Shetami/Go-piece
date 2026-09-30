WITH b AS (
  SELECT width_bucket(amount, 0, 5000, 5) AS bucket, count(*) AS cnt
  FROM orders
  WHERE status = 'paid' AND amount IS NOT NULL
  GROUP BY 1
)
SELECT (g - 1) * 1000                   AS lo,
       CASE WHEN g <= 5 THEN g * 1000 END AS hi,
       coalesce(b.cnt, 0)               AS orders,
       round(100.0 * coalesce(b.cnt, 0) / sum(coalesce(b.cnt, 0)) OVER (), 1) AS share_pct
FROM generate_series(1, 6) AS g
LEFT JOIN b ON b.bucket = g
ORDER BY lo;
