-- Чарт сентября 2024 по жанрам: первые два места (dense_rank по числу уникальных слушателей).
-- Прослушивание засчитывается от 30 секунд. Жанр NULL → 'без жанра'.
-- Колонки: genre, place, title, listeners. Порядок: genre, place, title.
SELECT genre, place, title, listeners
FROM (
  SELECT t.genre, t.title, count(*) AS listeners,
         row_number() OVER (PARTITION BY t.genre ORDER BY count(*) DESC) AS place
  FROM plays p
  JOIN tracks t ON t.id = p.track_id
  WHERE p.played_at BETWEEN '2024-09-01' AND '2024-09-30'
  GROUP BY t.genre, t.title
) r
WHERE place <= 2
ORDER BY genre, place, title;
