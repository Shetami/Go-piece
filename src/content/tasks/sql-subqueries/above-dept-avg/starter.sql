-- Сотрудники, чья зарплата строго выше средней по их отделу.
-- Колонки: name, dept, salary. Порядок: по dept, затем по name.
SELECT name, dept, salary
FROM employees
WHERE salary > (SELECT avg(salary) FROM employees)
ORDER BY dept, name;
