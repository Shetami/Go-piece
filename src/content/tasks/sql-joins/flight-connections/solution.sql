SELECT t.passenger,
       i.flight_no AS inbound,
       c.flight_no AS connection,
       (extract(epoch FROM c.dep_at - i.arr_at) / 60)::int AS wait_minutes
FROM transfers t
JOIN flights i ON i.id = t.inbound_flight_id
LEFT JOIN LATERAL (
  SELECT f.flight_no, f.dep_at
  FROM flights f
  WHERE f.origin = i.dest
    AND f.dest = t.final_dest
    AND f.status = 'scheduled'
    AND f.dep_at >= i.arr_at + interval '50 minutes'
    AND f.dep_at <= i.arr_at + interval '12 hours'
  ORDER BY f.dep_at, f.flight_no
  LIMIT 1
) c ON true
ORDER BY t.id;
