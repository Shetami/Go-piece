WITH slots AS (
  SELECT h, date '2024-08-05' + h * interval '1 hour' AS slot_start
  FROM generate_series(8, 19) AS h
),
grid AS (
  SELECT p.id AS point_id, p.name, s.h, s.slot_start,
         coalesce(own.required, dflt.required) AS required
  FROM points p
  CROSS JOIN slots s
  LEFT JOIN coverage_rules own
    ON own.point_id = p.id AND s.h >= own.hour_from AND s.h < own.hour_to
  LEFT JOIN coverage_rules dflt
    ON dflt.point_id IS NULL AND s.h >= dflt.hour_from AND s.h < dflt.hour_to
  WHERE p.is_active
),
staffed AS (
  SELECT g.point_id, g.name, g.slot_start, g.required,
         count(DISTINCT sh.employee_id) AS staffed
  FROM grid g
  LEFT JOIN shifts sh
    ON sh.point_id = g.point_id
   AND sh.starts_at <= g.slot_start
   AND sh.ends_at >= g.slot_start + interval '1 hour'
  GROUP BY g.point_id, g.name, g.slot_start, g.required
)
SELECT name AS point, to_char(slot_start, 'HH24:MI') AS slot, staffed, required
FROM staffed
WHERE staffed < required
ORDER BY name, slot_start;
