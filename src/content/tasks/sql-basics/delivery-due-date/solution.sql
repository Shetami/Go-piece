WITH base AS (
  SELECT id, sla_days, delivered_at,
         -- принятые с 18:00 считаются принятыми на следующий день
         accepted_at::date + CASE WHEN accepted_at::time >= time '18:00' THEN 1 ELSE 0 END AS start_day
  FROM shipments
),
due AS (
  SELECT b.*,
         (SELECT d::date
          FROM generate_series(b.start_day + 1, b.start_day + 60, interval '1 day') AS d
          WHERE extract(isodow FROM d) < 6
            AND NOT EXISTS (SELECT 1 FROM holidays h WHERE h.day = d::date)
          ORDER BY d
          OFFSET b.sla_days - 1 LIMIT 1) AS due_date   -- N-й рабочий день после старта
  FROM base b
)
SELECT id, due_date, delivered_at::date AS delivered_on,
       CASE WHEN delivered_at IS NULL THEN
                 CASE WHEN due_date >= date '2024-05-13' THEN 'в пути' ELSE 'просрочено' END
            WHEN delivered_at::date <= due_date THEN 'в срок'   -- сравниваем даты, а не timestamp с полночью
            ELSE 'опоздание' END AS status
FROM due
ORDER BY due_date, id;
