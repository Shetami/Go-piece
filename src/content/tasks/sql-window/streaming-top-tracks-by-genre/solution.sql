WITH listeners AS (
  SELECT t.id, t.title,
         coalesce(t.genre, 'без жанра') AS genre,
         count(DISTINCT p.user_id)      AS listeners
  FROM plays p
  JOIN tracks t ON t.id = p.track_id
  WHERE p.seconds >= 30
    AND p.played_at >= '2024-09-01'
    AND p.played_at <  '2024-10-01'
  GROUP BY t.id, t.title, t.genre
),
ranked AS (
  SELECT genre, title, listeners,
         dense_rank() OVER (PARTITION BY genre ORDER BY listeners DESC) AS place
  FROM listeners
)
SELECT genre, place, title, listeners
FROM ranked
WHERE place <= 2
ORDER BY genre, place, title;
