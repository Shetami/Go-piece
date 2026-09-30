WITH compared AS (
  SELECT room, sensor, ts, temp,
         avg(temp) OVER (
           PARTITION BY room, ts
           ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING
           EXCLUDE CURRENT ROW
         ) AS others_avg
  FROM readings
)
SELECT room, sensor, to_char(ts, 'HH24:MI') AS ts, temp,
       round(others_avg, 2) AS others_avg
FROM compared
WHERE abs(temp - others_avg) >= 5
ORDER BY room, compared.ts, sensor;
