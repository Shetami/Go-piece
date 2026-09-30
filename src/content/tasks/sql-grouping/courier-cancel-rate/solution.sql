WITH per AS (
  SELECT courier,
         count(*)                                     AS deliveries,
         count(*) FILTER (WHERE status = 'cancelled') AS cancelled,
         sum(count(*)) OVER ()                                     AS all_deliveries,
         sum(count(*) FILTER (WHERE status = 'cancelled')) OVER () AS all_cancelled
  FROM deliveries
  WHERE courier IS NOT NULL
  GROUP BY courier
)
SELECT courier, deliveries, cancelled,
       round(100.0 * cancelled / deliveries, 1) AS cancel_pct,
       round(100.0 * cancelled / deliveries - 100.0 * all_cancelled / all_deliveries, 1) AS vs_avg
FROM per
WHERE deliveries >= 5
ORDER BY cancel_pct DESC, courier;
