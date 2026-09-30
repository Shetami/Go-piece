WITH daily AS (
  SELECT delivered_at::date AS day, courier, count(*) AS done
  FROM deliveries
  WHERE status = 'done'
  GROUP BY 1, 2
)
SELECT day, courier, done,
       rank()     OVER by_day     AS day_rank,
       done - lag(done) OVER by_courier AS delta,
       sum(done)  OVER (by_month ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS month_total
FROM daily
WINDOW by_day     AS (PARTITION BY day ORDER BY done DESC),
       by_courier AS (PARTITION BY courier ORDER BY day),
       by_month   AS (PARTITION BY courier, date_trunc('month', day) ORDER BY day)
ORDER BY day, day_rank, courier;
