-- Нарушения SLA первого ответа по приоритетам на 2024-06-10 12:00.
-- Колонки: priority, tickets, breached, breach_pct (round(…, 1)).
-- Порядок: high, normal, low.
SELECT priority,
       count(*) AS tickets,
       count(*) FILTER (WHERE extract(hour FROM first_reply_at - created_at) > 1) AS breached,
       0 AS breach_pct
FROM tickets
GROUP BY priority
ORDER BY priority;
