SELECT e.name, e.salary, m.name AS manager_name, m.salary AS manager_salary
FROM employees e
JOIN employees m ON m.id = e.manager_id
WHERE e.salary > m.salary
ORDER BY e.name;
