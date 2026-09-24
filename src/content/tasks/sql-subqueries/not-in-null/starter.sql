-- Сотрудники, у которых нет подчинённых
SELECT name
FROM employees
WHERE id NOT IN (SELECT manager_id FROM employees);
