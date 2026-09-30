-- Для каждого устройства и каждого дня, когда от него были записи:
-- среднее за день и скользящее среднее дневных значений за этот и два предыдущих календарных дня.
-- NULL-показания не учитываются. Колонки: device, day, daily_avg, avg_3d (оба round 2).
-- Порядок: device, day.
WITH daily AS (
  SELECT device, ts::date AS day, avg(coalesce(temp, 0)) AS daily_avg
  FROM readings
  GROUP BY device, ts::date
)
SELECT device, day, round(daily_avg, 2),
       round(avg(daily_avg) OVER (PARTITION BY device ORDER BY day
                                  ROWS BETWEEN 2 PRECEDING AND CURRENT ROW), 2)
FROM daily
ORDER BY device, day;
