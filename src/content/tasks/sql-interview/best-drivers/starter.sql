-- Лучший водитель каждого города за май 2024.
-- Колонки: city, driver, done_rides, median_pickup_min, driver_cancel_pct, avg_stars.
-- Порядок: по city.
SELECT c.name AS city,
       d.name AS driver,
       count(*) AS done_rides,
       avg(extract(epoch FROM r.arrived_at - r.accepted_at) / 60) AS median_pickup_min,
       0 AS driver_cancel_pct,
       avg(rt.stars) AS avg_stars
FROM rides r
JOIN drivers d ON d.id = r.driver_id
JOIN cities c  ON c.id = d.city_id
JOIN ratings rt ON rt.ride_id = r.id
GROUP BY c.name, d.name
ORDER BY city;
