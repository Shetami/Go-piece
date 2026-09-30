SELECT coalesce(priority, 'none') AS priority,
       count(*)           AS tickets,
       count(resolved_at) AS resolved,
       round(avg(extract(epoch FROM resolved_at - created_at) / 3600), 1) AS avg_hours,
       round(avg(csat), 2) AS avg_csat
FROM tickets
GROUP BY priority
ORDER BY 1;
