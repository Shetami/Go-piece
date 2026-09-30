-- Маршрут посылки одной строкой: города в порядке первого появления, каждый один раз.
-- Колонки: tracking, cities (сколько разных городов), route ('—', если городов нет).
-- Порядок: по tracking.
SELECT p.tracking,
       count(s.city) AS cities,
       string_agg(s.city, ' → ' ORDER BY s.id) AS route
FROM parcels p
JOIN scans s ON s.parcel_id = p.id
GROUP BY p.tracking
ORDER BY p.tracking;
