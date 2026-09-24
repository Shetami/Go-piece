-- Все подчинённые Бориса (id = 2) — прямые и через любое число ступеней.
-- Колонки: name, depth (1 — прямой подчинённый, 2 — подчинённый подчинённого…).
-- Порядок: по depth, затем по name.
SELECT name, 1 AS depth
FROM employees
WHERE manager_id = 2
ORDER BY depth, name;
