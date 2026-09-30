WITH flagged AS (
  SELECT ticker, day,
         coalesce(close > lag(close) OVER (PARTITION BY ticker ORDER BY day), false) AS rising
  FROM closes
),
grouped AS (
  SELECT ticker, day, rising,
         count(*) FILTER (WHERE NOT rising) OVER (
           PARTITION BY ticker ORDER BY day
           ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
         ) AS grp
  FROM flagged
),
streaks AS (
  SELECT ticker, count(*) AS streak_days, min(day) AS from_day, max(day) AS to_day
  FROM grouped
  WHERE rising
  GROUP BY ticker, grp
),
best AS (
  SELECT *, row_number() OVER (PARTITION BY ticker ORDER BY streak_days DESC, from_day) AS rn
  FROM streaks
)
SELECT t.ticker,
       coalesce(b.streak_days, 0) AS streak_days,
       b.from_day, b.to_day
FROM (SELECT DISTINCT ticker FROM closes) t
LEFT JOIN best b ON b.ticker = t.ticker AND b.rn = 1
ORDER BY t.ticker;
