WITH billable AS (
  SELECT r.id, r.user_id, r.started_at,
         ceil(extract(epoch FROM r.ended_at - r.started_at) / 60)::int AS minutes
  FROM rides r
  WHERE r.ended_at IS NOT NULL
    AND r.ended_at - r.started_at >= interval '30 seconds'
),
minute_prices AS (
  -- каждая начатая минута тарифицируется по тарифу на момент её начала
  SELECT b.id, t.price_per_min
  FROM billable b
  CROSS JOIN LATERAL generate_series(0, b.minutes - 1) AS g(i)
  JOIN tariffs t ON CASE
                      WHEN t.from_time < t.to_time
                        THEN (b.started_at + g.i * interval '1 minute')::time >= t.from_time
                         AND (b.started_at + g.i * interval '1 minute')::time <  t.to_time
                      ELSE (b.started_at + g.i * interval '1 minute')::time >= t.from_time
                        OR (b.started_at + g.i * interval '1 minute')::time <  t.to_time
                    END
),
ride_cost AS (
  SELECT b.id, b.user_id, b.minutes,
         (SELECT sum(price_per_min) FROM minute_prices mp WHERE mp.id = b.id)
           + CASE WHEN u.has_subscription THEN 0 ELSE s.unlock_fee END AS cost
  FROM billable b
  JOIN users u ON u.id = b.user_id
  CROSS JOIN settings s
)
SELECT u.name AS user_name,
       count(rc.id)                 AS rides,
       coalesce(sum(rc.minutes), 0) AS minutes,
       coalesce(sum(rc.cost), 0)    AS amount
FROM users u
LEFT JOIN ride_cost rc ON rc.user_id = u.id
GROUP BY u.id, u.name
ORDER BY amount DESC, user_name;
