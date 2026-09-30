-- Аномалии: показание отличается от среднего по ДРУГИМ датчикам той же комнаты в тот же момент
-- на 5 градусов и больше. NULL-показания в среднем не участвуют и сами аномалией не бывают.
-- Колонки: room, sensor, ts ('HH24:MI'), temp, others_avg (round 2). Порядок: room, ts, sensor.
SELECT room, sensor, to_char(ts, 'HH24:MI') AS ts, temp, round(room_avg, 2) AS others_avg
FROM (
  SELECT r.*, avg(coalesce(temp, 0)) OVER (PARTITION BY room, ts) AS room_avg
  FROM readings r
) x
WHERE abs(temp - room_avg) >= 5
ORDER BY room, x.ts, sensor;
