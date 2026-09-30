-- Личные рекорды: пробежки, дистанция которых строго больше всех предыдущих пробежек бегуна
-- (порядок: run_date, затем id). Первая пробежка с известной дистанцией — рекорд.
-- Колонки: runner, run_date, distance_km, prev_best (прошлый рекорд, NULL у первого).
-- Порядок: runner, run_date, distance_km.
SELECT runner, run_date, distance_km, prev_best
FROM (
  SELECT runner, run_date, distance_km,
         max(distance_km) OVER (PARTITION BY runner ORDER BY run_date) AS best_so_far,
         lag(distance_km) OVER (PARTITION BY runner ORDER BY run_date) AS prev_best
  FROM runs
) x
WHERE distance_km = best_so_far
ORDER BY runner, run_date, distance_km;
