-- Простой машин между рейсами: от приезда до выезда в следующий рейс (рейсы по времени выезда).
-- Если следующий рейс начался раньше приезда — простой 0. Все машины, без рейсов — нули.
-- Колонки: truck, trips, idle_hours (round 2), longest_idle_min (int). Порядок: idle_hours по убыванию, truck.
WITH gaps AS (
  SELECT truck, departed_at - lag(arrived_at) OVER (PARTITION BY truck ORDER BY id) AS idle
  FROM trips
)
SELECT truck, count(*) AS trips,
       round(sum(extract(epoch FROM idle)) / 3600, 2) AS idle_hours,
       max(extract(epoch FROM idle))::int / 60 AS longest_idle_min
FROM gaps
GROUP BY truck
ORDER BY idle_hours DESC, truck;
