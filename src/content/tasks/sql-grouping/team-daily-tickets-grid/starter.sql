-- Нагрузка команд поддержки за 1–7 июля 2024.
-- Колонки: team, tickets, zero_days, busiest_day (date или NULL), busiest_count.
-- Порядок: по team.
SELECT t.name AS team,
       count(*) AS tickets,
       7 - count(DISTINCT k.created_at::date) AS zero_days,
       min(k.created_at::date) AS busiest_day,
       count(*) AS busiest_count
FROM teams t
JOIN tickets k ON k.team_id = t.id
WHERE k.created_at::date BETWEEN '2024-07-01' AND '2024-07-07'
GROUP BY t.name
ORDER BY t.name;
