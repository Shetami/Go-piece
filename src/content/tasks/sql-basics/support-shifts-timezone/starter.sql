-- Обращения за март 2024 по сменам (московское время), будни и выходные.
-- Колонки: shift, weekdays, weekends.
-- Порядок: ночь, день, вечер.
SELECT CASE WHEN extract(hour FROM created_at) < 8 THEN 'ночь'
            WHEN extract(hour FROM created_at) < 16 THEN 'день'
            ELSE 'вечер' END AS shift,
       count(*) FILTER (WHERE extract(isodow FROM created_at) < 6) AS weekdays,
       count(*) FILTER (WHERE extract(isodow FROM created_at) >= 6) AS weekends
FROM tickets
WHERE created_at BETWEEN '2024-03-01' AND '2024-03-31'
GROUP BY 1
ORDER BY 1;
