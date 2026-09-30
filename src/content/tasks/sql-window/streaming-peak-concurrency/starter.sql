-- Пик одновременных зрителей каждой трансляции и первый момент, когда он был достигнут.
-- Вышедший в момент T и зашедший в момент T одновременно не смотрят.
-- ended_at IS NULL — смотрел до конца трансляции. Трансляции без зрителей: 0 и NULL.
-- Колонки: title, peak, peak_at ('HH24:MI'). Порядок: title.
WITH events AS (
  SELECT stream_id, started_at AS ts, 1 AS delta FROM views
  UNION ALL
  SELECT stream_id, ended_at, -1 FROM views WHERE ended_at IS NOT NULL
),
running AS (
  SELECT stream_id, ts, sum(delta) OVER (PARTITION BY stream_id ORDER BY ts, delta DESC
                                         ROWS UNBOUNDED PRECEDING) AS online
  FROM events
)
SELECT s.title, max(r.online) AS peak, to_char(min(r.ts), 'HH24:MI') AS peak_at
FROM streams s
JOIN running r ON r.stream_id = s.id
GROUP BY s.title
ORDER BY s.title;
