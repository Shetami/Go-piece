-- Сколько посылок сейчас в каждом статусе (статус — по последнему событию).
-- Колонки: status, parcels, ids (id посылок через ', ' по возрастанию).
-- Порядок: parcels по убыванию, затем status.
SELECT status,
       count(*) AS parcels,
       string_agg(parcel_id::text, ', ') AS ids
FROM parcel_events
WHERE id IN (SELECT max(id) FROM parcel_events GROUP BY parcel_id)
GROUP BY status
ORDER BY parcels DESC, status;
