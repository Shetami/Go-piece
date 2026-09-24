SELECT d.name, count(*) AS staff
FROM departments d
LEFT JOIN employees e ON e.dept_id = d.id
GROUP BY d.name
ORDER BY d.name;
