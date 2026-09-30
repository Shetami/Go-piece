-- SLA первого ответа в рабочих минутах по командам поддержки.
-- Колонки: team, tickets, answered, breached, avg_reply_min. Порядок: по team.
SELECT tm.name AS team,
       count(*) AS tickets,
       count(m.created_at) AS answered,
       count(*) FILTER (WHERE m.created_at - t.created_at > interval '2 hours') AS breached,
       round(avg(extract(epoch FROM m.created_at - t.created_at) / 60), 1) AS avg_reply_min
FROM teams tm
JOIN tickets t       ON t.team_id = tm.id
LEFT JOIN messages m ON m.ticket_id = t.id
GROUP BY tm.name
ORDER BY team;
