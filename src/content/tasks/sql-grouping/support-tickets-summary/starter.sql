-- Сводка по приоритетам тикетов.
-- Колонки: priority (NULL -> 'none'), tickets, resolved, avg_hours (round 1),
--          avg_csat (round 2). Порядок: по priority.
SELECT coalesce(priority, 'none') AS priority,
       count(*) AS tickets,
       count(*) AS resolved,
       round(avg(coalesce(extract(epoch FROM resolved_at - created_at) / 3600, 0)), 1) AS avg_hours,
       round(avg(coalesce(csat, 0)), 2) AS avg_csat
FROM tickets
GROUP BY priority
ORDER BY 1;
