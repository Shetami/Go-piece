WITH free AS (
  SELECT e.id AS event_id,
         count(r.id) FILTER (WHERE r.status = 'confirmed') AS confirmed,
         e.capacity - count(r.id) FILTER (WHERE r.status = 'confirmed') AS seats
  FROM events e
  LEFT JOIN registrations r ON r.event_id = e.id
  GROUP BY e.id, e.capacity
),
queue AS (
  SELECT id, event_id,
         row_number() OVER (PARTITION BY event_id ORDER BY created_at, id) AS pos
  FROM registrations
  WHERE status = 'waitlist'
),
promoted AS (
  UPDATE registrations r
  SET status = 'confirmed'
  FROM queue q
  JOIN free f ON f.event_id = q.event_id
  WHERE r.id = q.id
    AND q.pos <= f.seats
  RETURNING r.event_id, r.person, q.pos, f.confirmed
)
SELECT e.title,
       p.person,
       p.pos AS position,
       p.confirmed + count(*) OVER (PARTITION BY p.event_id) AS confirmed_after
FROM promoted p
JOIN events e ON e.id = p.event_id
ORDER BY e.title, p.pos;
