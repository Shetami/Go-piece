WITH may AS (
  SELECT *
  FROM rides
  WHERE accepted_at >= '2024-05-01' AND accepted_at < '2024-06-01'
),
stats AS (
  SELECT d.id, d.name, d.city_id,
         count(*) FILTER (WHERE m.status = 'done') AS done_rides,
         percentile_cont(0.5) WITHIN GROUP (
           ORDER BY extract(epoch FROM m.arrived_at - m.accepted_at) / 60
         ) FILTER (WHERE m.status = 'done') AS median_pickup,
         100.0 * count(*) FILTER (WHERE m.status = 'cancelled_by_driver') / count(*) AS driver_cancel_pct,
         avg(r.stars) AS avg_stars
  FROM drivers d
  JOIN may m          ON m.driver_id = d.id
  LEFT JOIN ratings r ON r.ride_id = m.id
  GROUP BY d.id, d.name, d.city_id
  HAVING count(*) FILTER (WHERE m.status = 'done') >= 3
),
ranked AS (
  SELECT *,
         row_number() OVER (PARTITION BY city_id
                            ORDER BY median_pickup, avg_stars DESC NULLS LAST, name) AS rn
  FROM stats
)
SELECT c.name AS city,
       r.name AS driver,
       r.done_rides,
       round(r.median_pickup::numeric, 1) AS median_pickup_min,
       round(r.driver_cancel_pct, 1)      AS driver_cancel_pct,
       round(r.avg_stars, 2)              AS avg_stars
FROM ranked r
JOIN cities c ON c.id = r.city_id
WHERE r.rn = 1
ORDER BY city;
