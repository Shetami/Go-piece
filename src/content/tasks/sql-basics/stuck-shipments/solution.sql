WITH last_event AS (
  SELECT DISTINCT ON (shipment_id)
         shipment_id, lower(trim(status)) AS status, happened_at
  FROM shipment_events
  ORDER BY shipment_id, happened_at DESC, id DESC   -- при равном времени верим более позднему id
),
cur AS (
  SELECT s.id                                 AS shipment_id,
         coalesce(e.status, 'created')        AS status,
         coalesce(e.happened_at, s.created_at) AS last_at
  FROM shipments s
  LEFT JOIN last_event e ON e.shipment_id = s.id
)
SELECT shipment_id, status, last_at,
       floor(extract(epoch FROM timestamp '2024-07-01 12:00' - last_at) / 3600)::int AS hours
FROM cur
WHERE status NOT IN ('delivered', 'cancelled')
  AND last_at < timestamp '2024-07-01 12:00' - interval '48 hours'
ORDER BY hours DESC, shipment_id;
