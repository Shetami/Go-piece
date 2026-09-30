-- Ближайшая стыковка для каждого пассажира (не раньше 50 минут и не позже 12 часов после прилёта).
-- Колонки: passenger, inbound, connection (NULL, если стыковки нет), wait_minutes.
-- Порядок: по id заявки на трансфер.
SELECT t.passenger, i.flight_no AS inbound, min(f.flight_no) AS connection,
       (extract(epoch FROM min(f.dep_at) - i.arr_at) / 60)::int AS wait_minutes
FROM transfers t
JOIN flights i ON i.id = t.inbound_flight_id
JOIN flights f ON f.dest = t.final_dest AND f.dep_at > i.arr_at
GROUP BY t.id, t.passenger, i.flight_no, i.arr_at
ORDER BY t.id;
