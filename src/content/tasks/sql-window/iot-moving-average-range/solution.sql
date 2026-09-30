WITH daily AS (
  SELECT device, ts::date AS day, avg(temp) AS daily_avg
  FROM readings
  GROUP BY device, ts::date
)
SELECT device, day,
       round(daily_avg, 2) AS daily_avg,
       round(avg(daily_avg) OVER (
         PARTITION BY device
         ORDER BY day
         RANGE BETWEEN INTERVAL '2 days' PRECEDING AND CURRENT ROW
       ), 2) AS avg_3d
FROM daily
ORDER BY device, day;
