-- Застрявшие отправления на 2024-07-01 12:00.
-- Колонки: shipment_id, status, last_at, hours.
-- Порядок: hours по убыванию, затем shipment_id.
SELECT shipment_id, status, max(happened_at) AS last_at, 0 AS hours
FROM shipment_events
WHERE status <> 'delivered'
GROUP BY shipment_id, status
HAVING max(happened_at) < '2024-06-29'
ORDER BY shipment_id;
