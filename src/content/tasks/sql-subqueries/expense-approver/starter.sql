-- Кто согласует каждый расход: ближайший руководитель вверх по цепочке (не сам автор),
-- у которого approval_limit >= amount. Нет такого — approver и step пустые.
-- Колонки: id, employee, approver, step (1 — непосредственный руководитель). Порядок: id.
SELECT x.id, e.name AS employee, m.name AS approver, 1 AS step
FROM expenses x
JOIN employees e      ON e.id = x.employee_id
LEFT JOIN employees m ON m.id = e.manager_id
ORDER BY x.id;
