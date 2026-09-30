WITH legs AS (
  SELECT parcel,
         hub AS to_hub,
         lag(hub)        OVER w AS from_hub,
         scanned_at - lag(scanned_at) OVER w AS dur,
         scanned_at
  FROM scans
  WINDOW w AS (PARTITION BY parcel ORDER BY scanned_at, id)
),
ranked AS (
  SELECT parcel, from_hub, to_hub,
         (extract(epoch FROM dur) / 60)::int AS minutes,
         row_number() OVER (PARTITION BY parcel ORDER BY dur DESC, scanned_at) AS rn
  FROM legs
  WHERE from_hub IS NOT NULL
)
SELECT p.code AS parcel, r.from_hub, r.to_hub, r.minutes
FROM parcels p
LEFT JOIN ranked r ON r.parcel = p.code AND r.rn = 1
ORDER BY p.code;
