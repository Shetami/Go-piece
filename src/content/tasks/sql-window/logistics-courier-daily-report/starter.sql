-- Дневной отчёт по курьерам (только успешные доставки, только дни, когда у курьера они были):
-- done, место за день (rank, больше — выше), разница с предыдущим рабочим днём курьера,
-- нарастающий итог курьера с начала месяца.
-- Колонки: day, courier, done, day_rank, delta, month_total. Порядок: day, day_rank, courier.
WITH daily AS (
  SELECT delivered_at::date AS day, courier, count(*) AS done
  FROM deliveries
  GROUP BY 1, 2
)
SELECT day, courier, done,
       row_number() OVER (PARTITION BY day ORDER BY done DESC) AS day_rank,
       done - lag(done) OVER (ORDER BY day) AS delta,
       sum(done) OVER (PARTITION BY courier ORDER BY day) AS month_total
FROM daily
ORDER BY day, day_rank, courier;
