WITH current AS (
  SELECT parcel_id,
         (array_agg(status ORDER BY happened_at DESC, id DESC))[1] AS status
  FROM parcel_events
  GROUP BY parcel_id
)
SELECT status,
       count(*) AS parcels,
       string_agg(parcel_id::text, ', ' ORDER BY parcel_id) AS ids
FROM current
GROUP BY status
ORDER BY parcels DESC, status;
