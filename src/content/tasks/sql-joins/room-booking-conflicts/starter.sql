-- Пары пересекающихся броней одной переговорки (отменённые не считаются).
-- Колонки: room, first_id, second_id (first_id < second_id), overlap_minutes.
-- Порядок: по room, first_id, second_id.
SELECT r.name AS room, a.id AS first_id, b.id AS second_id,
       (extract(epoch FROM a.ends_at - b.starts_at) / 60)::int AS overlap_minutes
FROM bookings a
JOIN bookings b ON b.room_id = a.room_id AND a.id <> b.id
               AND b.starts_at BETWEEN a.starts_at AND a.ends_at
JOIN rooms r ON r.id = a.room_id
ORDER BY r.name, a.id, b.id;
