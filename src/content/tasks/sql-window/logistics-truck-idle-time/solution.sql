WITH gaps AS (
  SELECT truck,
         greatest(lead(departed_at) OVER (PARTITION BY truck ORDER BY departed_at, id) - arrived_at,
                  INTERVAL '0') AS idle
  FROM trips
)
SELECT t.plate AS truck,
       count(g.truck) AS trips,
       round(coalesce(sum(extract(epoch FROM g.idle)), 0) / 3600, 2) AS idle_hours,
       coalesce(max(extract(epoch FROM g.idle)), 0)::int / 60 AS longest_idle_min
FROM trucks t
LEFT JOIN gaps g ON g.truck = t.plate
GROUP BY t.plate
ORDER BY idle_hours DESC, t.plate;
