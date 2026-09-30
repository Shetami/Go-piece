SELECT d.day::date AS day,
       count(DISTINCT s.user_id)                             AS active_users,
       count(DISTINCT s.user_id) FILTER (WHERE s.plan = 'pro') AS pro_users
FROM generate_series(date '2024-06-01', date '2024-06-07', interval '1 day') AS d(day)
LEFT JOIN subscriptions s
       ON s.started_at <= d.day
      AND (s.ended_at IS NULL OR s.ended_at > d.day)
GROUP BY d.day
ORDER BY d.day;
