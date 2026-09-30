-- Часы 5 августа 2024 (с 08:00 до 20:00), когда в пункте выдачи меньше людей, чем нужно.
-- Колонки: point, slot ('HH24:MI' — начало часа), staffed, required.
-- Порядок: по point, затем по slot.
SELECT p.name AS point, to_char(sh.starts_at, 'HH24:MI') AS slot,
       count(*) AS staffed, r.required
FROM points p
JOIN shifts sh ON sh.point_id = p.id
JOIN coverage_rules r ON r.point_id IS NULL
                     AND extract(hour FROM sh.starts_at) BETWEEN r.hour_from AND r.hour_to
GROUP BY p.name, sh.starts_at, r.required
HAVING count(*) < r.required
ORDER BY point, slot;
