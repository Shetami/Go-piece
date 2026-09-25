WITH nights AS (
  SELECT r.room_type_id, g.discount_pct, d::date AS night
  FROM bookings b
  JOIN rooms r  ON r.id = b.room_id
  JOIN guests g ON g.id = b.guest_id
  CROSS JOIN LATERAL generate_series(b.check_in, b.check_out - 1, interval '1 day') AS d
  WHERE b.status = 'confirmed'
    AND d >= date '2024-03-01'
    AND d <  date '2024-04-01'
),
sold AS (
  SELECT n.room_type_id,
         count(*) AS nights,
         sum(t.price_per_night * (100 - n.discount_pct) / 100.0) AS revenue
  FROM nights n
  JOIN tariffs t ON t.room_type_id = n.room_type_id
                AND n.night BETWEEN t.valid_from AND t.valid_to
  GROUP BY n.room_type_id
),
capacity AS (
  SELECT room_type_id, count(*) AS rooms
  FROM rooms
  GROUP BY room_type_id
)
SELECT rt.name AS room_type,
       coalesce(s.nights, 0) AS nights,
       round(100.0 * coalesce(s.nights, 0) / (c.rooms * 31), 1) AS occupancy_pct,
       round(coalesce(s.revenue, 0), 2) AS revenue
FROM room_types rt
JOIN capacity c ON c.room_type_id = rt.id
LEFT JOIN sold s ON s.room_type_id = rt.id
ORDER BY revenue DESC, room_type;
