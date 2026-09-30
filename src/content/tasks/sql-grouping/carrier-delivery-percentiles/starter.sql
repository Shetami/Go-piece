-- Сроки доставки по перевозчикам.
-- Колонки: carrier, shipments, delivered, median_h, p90_h (часы, round 1).
-- Порядок: по carrier.
SELECT carrier,
       count(*) AS shipments,
       count(*) AS delivered,
       round(avg(extract(epoch FROM delivered_at - shipped_at) / 3600), 1) AS median_h,
       round(max(extract(epoch FROM delivered_at - shipped_at) / 3600), 1) AS p90_h
FROM shipments
WHERE delivered_at IS NOT NULL
GROUP BY carrier
ORDER BY carrier;
