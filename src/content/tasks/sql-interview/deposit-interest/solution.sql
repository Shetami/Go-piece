WITH days AS (
  SELECT d::date AS day
  FROM generate_series(date '2024-10-01', date '2024-10-31', interval '1 day') AS d
),
daily AS (
  SELECT a.id, a.client, d.day,
         coalesce((SELECT sum(m.amount) FROM movements m
                   WHERE m.account_id = a.id AND m.op_date <= d.day), 0) AS balance,
         (SELECT r.rate_pct FROM rates r
          WHERE r.valid_from <= d.day
          ORDER BY r.valid_from DESC
          LIMIT 1) AS rate_pct
  FROM accounts a
  CROSS JOIN days d
)
SELECT client,
       round(avg(balance), 2) AS avg_balance,
       round(sum(greatest(balance, 0) * rate_pct / 100 / 366), 2) AS interest
FROM daily
GROUP BY id, client
ORDER BY interest DESC, client;
