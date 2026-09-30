-- Первые 4 свободных места подряд в одном ряду на каждый сеанс.
-- Колонки: film, starts_at, row_no, first_seat, last_seat. Порядок: starts_at, film.
SELECT sc.film, sc.starts_at, s.row_no,
       min(s.seat_no) AS first_seat, min(s.seat_no) + 3 AS last_seat
FROM screenings sc
JOIN seats s ON s.hall_id = sc.hall_id
WHERE s.id NOT IN (SELECT seat_id FROM tickets)
GROUP BY sc.film, sc.starts_at, s.row_no
HAVING count(*) >= 4
ORDER BY sc.starts_at, sc.film;
