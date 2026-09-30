-- Оргструктура в порядке обхода в глубину (дети одного руководителя — по имени).
-- Колонки: name, level (верхний уровень = 1), path ('Анна / Борис / Вика'),
-- team_size (все подчинённые на всех уровнях). Порядок: как в дереве.
SELECT e.name,
       CASE WHEN e.manager_id IS NULL THEN 1 ELSE 2 END AS level,
       coalesce(m.name || ' / ', '') || e.name AS path,
       (SELECT count(*) FROM employees s WHERE s.manager_id = e.id) AS team_size
FROM employees e
LEFT JOIN employees m ON m.id = e.manager_id
ORDER BY path;
