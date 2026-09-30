-- Перевести из листа ожидания в подтверждённые столько людей, сколько свободных мест
-- (capacity - confirmed), по очереди created_at, затем id. Вернуть переведённых:
-- title, person, position (место в очереди), confirmed_after (подтверждённых после). Порядок: title, position.
WITH promoted AS (
  UPDATE registrations r
  SET status = 'confirmed'
  WHERE r.status = 'waitlist'
    AND (SELECT count(*) FROM registrations c WHERE c.event_id = r.event_id AND c.status <> 'waitlist')
        < (SELECT capacity FROM events e WHERE e.id = r.event_id)
  RETURNING r.event_id, r.person, r.created_at
)
SELECT e.title,
       p.person,
       row_number() OVER (PARTITION BY p.event_id ORDER BY p.created_at) AS position,
       (SELECT count(*) FROM registrations c WHERE c.event_id = p.event_id AND c.status = 'confirmed') AS confirmed_after
FROM promoted p
JOIN events e ON e.id = p.event_id
ORDER BY e.title, position;
