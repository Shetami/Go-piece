SELECT r.name AS room,
       a.id   AS first_id,
       b.id   AS second_id,
       (extract(epoch FROM least(a.ends_at, b.ends_at)
                         - greatest(a.starts_at, b.starts_at)) / 60)::int AS overlap_minutes
FROM bookings a
JOIN bookings b
  ON b.room_id = a.room_id
 AND a.id < b.id
 AND a.starts_at < b.ends_at
 AND b.starts_at < a.ends_at
JOIN rooms r ON r.id = a.room_id
WHERE a.status <> 'cancelled'
  AND b.status <> 'cancelled'
ORDER BY r.name, a.id, b.id;
