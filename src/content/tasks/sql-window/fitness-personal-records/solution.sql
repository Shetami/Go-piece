WITH with_best AS (
  SELECT runner, run_date, id, distance_km,
         max(distance_km) OVER (
           PARTITION BY runner
           ORDER BY run_date, id
           ROWS BETWEEN UNBOUNDED PRECEDING AND 1 PRECEDING
         ) AS prev_best
  FROM runs
)
SELECT runner, run_date, distance_km, prev_best
FROM with_best
WHERE distance_km IS NOT NULL
  AND (prev_best IS NULL OR distance_km > prev_best)
ORDER BY runner, run_date, distance_km;
