-- Топ-3 операторов по решённым тикетам за март 2024 (с разделившими 3-е место).
-- Колонки: name, resolved.
-- Порядок: resolved по убыванию, затем name.
SELECT a.name, count(*) AS resolved
FROM tickets t
JOIN agents a ON a.id = t.agent_id
WHERE t.status = 'resolved'
  AND t.closed_at BETWEEN '2024-03-01' AND '2024-03-31'
GROUP BY a.name
ORDER BY resolved DESC, a.name
LIMIT 3;
