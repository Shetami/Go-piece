-- Схлопнуть статусы в непрерывные интервалы. Интервал длится до начала следующего
-- интервала устройства, последний — до 12:00 (момент отчёта).
-- Колонки: device, status, started_at, ended_at (обе 'HH24:MI'), minutes (int).
-- Порядок: device, started_at.
SELECT device, status,
       to_char(ts, 'HH24:MI') AS started_at,
       to_char(lead(ts) OVER (PARTITION BY device ORDER BY ts), 'HH24:MI') AS ended_at,
       (extract(epoch FROM lead(ts) OVER (PARTITION BY device ORDER BY ts) - ts) / 60)::int AS minutes
FROM device_status
ORDER BY device, ts;
