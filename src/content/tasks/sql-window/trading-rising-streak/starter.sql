-- Самая длинная серия роста: подряд идущие торговые дни, в каждый из которых close строго
-- выше, чем в предыдущий торговый день. Длина — число таких дней; при равной длине — более ранняя.
-- Колонки: ticker, streak_days, from_day, to_day (первый и последний день роста). Без роста: 0, NULL, NULL.
-- Порядок: ticker.
WITH rising AS (
  SELECT ticker, day
  FROM (SELECT ticker, day, close, lag(close) OVER (PARTITION BY ticker ORDER BY day) AS prev
        FROM closes) x
  WHERE close >= prev
),
grouped AS (
  SELECT ticker, day, day - row_number() OVER (PARTITION BY ticker ORDER BY day)::int AS grp
  FROM rising
)
SELECT ticker, max(cnt) AS streak_days, min(from_day) AS from_day, max(to_day) AS to_day
FROM (SELECT ticker, count(*) AS cnt, min(day) AS from_day, max(day) AS to_day
      FROM grouped GROUP BY ticker, grp) s
GROUP BY ticker
ORDER BY ticker;
