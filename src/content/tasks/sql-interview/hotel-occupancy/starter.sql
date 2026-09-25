-- Отчёт по типам номеров за март 2024.
-- Колонки: room_type, nights, occupancy_pct, revenue.
-- Порядок: revenue по убыванию, затем room_type.
SELECT rt.name AS room_type,
       sum(b.check_out - b.check_in) AS nights,
       round(100.0 * sum(b.check_out - b.check_in) / (count(r.id) * 31), 1) AS occupancy_pct,
       sum((b.check_out - b.check_in) * t.price_per_night) AS revenue
FROM room_types rt
JOIN rooms r    ON r.room_type_id = rt.id
JOIN bookings b ON b.room_id = r.id
JOIN tariffs t  ON t.room_type_id = rt.id AND b.check_in BETWEEN t.valid_from AND t.valid_to
WHERE b.check_in >= '2024-03-01' AND b.check_in < '2024-04-01'
GROUP BY rt.name
ORDER BY revenue DESC, room_type;
