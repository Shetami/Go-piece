WITH taken AS (
  SELECT screening_id, seat_id
  FROM tickets
  WHERE status = 'sold'
     OR (status = 'reserved' AND reserved_until > timestamp '2024-08-01 18:00')
),
free AS (
  SELECT sc.id AS screening_id, s.row_no, s.seat_no,
         s.seat_no - row_number() OVER (PARTITION BY sc.id, s.row_no ORDER BY s.seat_no) AS grp
  FROM screenings sc
  JOIN seats s ON s.hall_id = sc.hall_id
  WHERE NOT s.is_broken
    AND NOT EXISTS (SELECT 1 FROM taken t
                    WHERE t.screening_id = sc.id AND t.seat_id = s.id)
),
blocks AS (
  SELECT screening_id, row_no, min(seat_no) AS first_seat
  FROM free
  GROUP BY screening_id, row_no, grp
  HAVING count(*) >= 4
),
best AS (
  SELECT DISTINCT ON (screening_id) screening_id, row_no, first_seat
  FROM blocks
  ORDER BY screening_id, row_no, first_seat
)
SELECT sc.film, sc.starts_at, b.row_no, b.first_seat, b.first_seat + 3 AS last_seat
FROM screenings sc
LEFT JOIN best b ON b.screening_id = sc.id
ORDER BY sc.starts_at, sc.film;
