WITH t AS (
  SELECT carrier, delivered_at,
         extract(epoch FROM delivered_at - shipped_at) / 3600 AS hours
  FROM shipments
)
SELECT carrier,
       count(*)            AS shipments,
       count(delivered_at) AS delivered,
       round(percentile_cont(0.5) WITHIN GROUP (ORDER BY hours)::numeric, 1) AS median_h,
       round(percentile_cont(0.9) WITHIN GROUP (ORDER BY hours)::numeric, 1) AS p90_h
FROM t
GROUP BY carrier
ORDER BY carrier;
