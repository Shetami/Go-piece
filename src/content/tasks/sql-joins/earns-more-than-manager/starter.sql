-- Сотрудники, которые получают больше своего непосредственного руководителя.
-- Колонки: name, salary, manager_name, manager_salary. Порядок: по name.
SELECT e.name, e.salary
FROM employees e
ORDER BY e.name;
