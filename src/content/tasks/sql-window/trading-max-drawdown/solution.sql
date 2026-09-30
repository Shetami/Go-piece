WITH peaks AS (
  SELECT account, day, balance,
         max(balance) OVER (PARTITION BY account ORDER BY day
                            ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS peak
  FROM equity
),
dd AS (
  SELECT account, day, balance, peak,
         min(day) OVER (PARTITION BY account, peak) AS peak_day,
         (peak - balance) / peak AS drawdown
  FROM peaks
),
worst AS (
  SELECT account, drawdown, peak_day, day AS trough_day,
         row_number() OVER (PARTITION BY account ORDER BY drawdown DESC, day) AS rn
  FROM dd
)
SELECT account,
       round(100 * drawdown, 2) AS max_drawdown_pct,
       CASE WHEN drawdown > 0 THEN peak_day END   AS peak_day,
       CASE WHEN drawdown > 0 THEN trough_day END AS trough_day
FROM worst
WHERE rn = 1
ORDER BY account;
