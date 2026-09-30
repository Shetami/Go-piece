WITH first_reply AS (
  SELECT ticket_id, min(created_at) AS replied_at
  FROM messages
  WHERE author = 'agent'
  GROUP BY ticket_id
),
spans AS (
  -- без ответа время идёт до момента отчёта
  SELECT t.id, t.team_id, t.priority, t.created_at AS started_at,
         coalesce(f.replied_at, timestamp '2024-07-08 12:00') AS ended_at,
         f.replied_at IS NOT NULL AS answered
  FROM tickets t
  LEFT JOIN first_reply f ON f.ticket_id = t.id
),
worked AS (
  SELECT s.id,
         coalesce(sum(greatest(0, extract(epoch FROM
                    least(s.ended_at, d + time '19:00') - greatest(s.started_at, d + time '10:00')
                  ) / 60)), 0) AS minutes
  FROM spans s
  CROSS JOIN LATERAL generate_series(s.started_at::date, s.ended_at::date, interval '1 day') AS g(d)
  WHERE extract(isodow FROM d) < 6
    AND d::date NOT IN (SELECT day FROM holidays)
  GROUP BY s.id
),
per_ticket AS (
  SELECT s.team_id, s.answered,
         coalesce(w.minutes, 0) AS minutes,
         coalesce(w.minutes, 0) > 60 * l.first_reply_h AS breached
  FROM spans s
  JOIN sla l ON l.priority = s.priority
  LEFT JOIN worked w ON w.id = s.id
)
SELECT tm.name AS team,
       count(p.team_id)                          AS tickets,
       count(*) FILTER (WHERE p.answered)        AS answered,
       count(*) FILTER (WHERE p.breached)        AS breached,
       round(avg(p.minutes) FILTER (WHERE p.answered), 1) AS avg_reply_min
FROM teams tm
LEFT JOIN per_ticket p ON p.team_id = tm.id
GROUP BY tm.id, tm.name
ORDER BY team;
