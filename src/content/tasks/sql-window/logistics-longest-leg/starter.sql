-- Самое долгое плечо каждой посылки: между двумя соседними сканами.
-- Сканы в одну минуту — по возрастанию id. При равной длительности — более раннее плечо.
-- Посылки меньше чем с двумя сканами — с NULL. Колонки: parcel, from_hub, to_hub, minutes (int).
-- Порядок: parcel.
SELECT parcel, from_hub, to_hub, minutes
FROM (
  SELECT parcel, lag(hub) OVER (PARTITION BY parcel ORDER BY id) AS from_hub, hub AS to_hub,
         extract(epoch FROM scanned_at::time - lag(scanned_at::time) OVER (PARTITION BY parcel ORDER BY id))::int / 60 AS minutes
  FROM scans
) l
WHERE minutes = (SELECT max(extract(epoch FROM s2.scanned_at - s1.scanned_at)::int / 60)
                 FROM scans s1 JOIN scans s2 ON s1.parcel = s2.parcel AND s2.id = s1.id + 1
                 WHERE s1.parcel = l.parcel)
ORDER BY parcel;
