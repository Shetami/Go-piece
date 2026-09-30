WITH first_seen AS (
  SELECT parcel_id, city, min(scanned_at) AS first_at
  FROM scans
  WHERE city IS NOT NULL
  GROUP BY parcel_id, city
)
SELECT p.tracking,
       count(f.city) AS cities,
       coalesce(string_agg(f.city, ' → ' ORDER BY f.first_at), '—') AS route
FROM parcels p
LEFT JOIN first_seen f ON f.parcel_id = p.id
GROUP BY p.id, p.tracking
ORDER BY p.tracking;
