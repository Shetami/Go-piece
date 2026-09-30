WITH m AS (
  SELECT date '2024-02-01'                                  AS m_start,
         (date '2024-02-01' + interval '1 month')::date - 1 AS m_end    -- последний день месяца, 29-е
),
x AS (
  SELECT s.customer, p.name AS plan, p.monthly_price, m.m_end - m.m_start + 1 AS month_days,
         -- least/greatest в Postgres пропускают NULL: действующая подписка упирается в конец месяца
         least(s.ended_on, m.m_end) - greatest(s.started_on, m.m_start) + 1 AS days
  FROM subscriptions s
  JOIN plans p ON p.id = s.plan_id
  CROSS JOIN m
)
SELECT customer, plan, days,
       round(monthly_price * days::numeric / month_days, 2) AS amount
FROM x
WHERE days > 0
ORDER BY customer, plan;
