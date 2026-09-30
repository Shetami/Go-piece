SELECT e.name, e.hired_on, a.d AS anniversary, a.n AS years
FROM employees e
CROSS JOIN LATERAL (
  -- годовщина = дата найма + n лет; 29 февраля в невисокосный год даёт 28-е
  SELECT n, (e.hired_on + n * interval '1 year')::date AS d
  FROM generate_series(1, 60) AS n
) a
WHERE a.d BETWEEN date '2024-12-20' AND date '2025-03-05'   -- обе границы — даты, BETWEEN уместен
  AND (e.fired_on IS NULL OR e.fired_on > a.d)
ORDER BY anniversary, e.name;
