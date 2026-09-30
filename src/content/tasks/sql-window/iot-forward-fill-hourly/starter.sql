-- Почасовая сетка 00:00–05:00 за 2024-02-01 для каждого устройства: последнее не-NULL показание
-- внутри часа; если в часе его нет — последнее известное из прошлых часов; до первого — NULL.
-- Колонки: device, hour ('HH24:MI'), value. Порядок: device, hour.
WITH hourly AS (
  SELECT device, date_trunc('hour', ts) AS hour, max(value) AS value
  FROM meter
  GROUP BY 1, 2
)
SELECT device, to_char(hour, 'HH24:MI') AS hour,
       coalesce(value, lag(value) OVER (PARTITION BY device ORDER BY hour)) AS value
FROM hourly
WHERE hour < '2024-02-01 06:00'
ORDER BY device, hourly.hour;
