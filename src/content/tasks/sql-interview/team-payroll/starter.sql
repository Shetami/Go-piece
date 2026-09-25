-- Руководители и их команды на 2024-06-01.
-- Колонки: manager, direct_reports, team_size, team_payroll, depth.
-- Порядок: team_size по убыванию, затем manager.
SELECT m.name AS manager,
       count(DISTINCT e.id) AS direct_reports,
       count(DISTINCT e.id) AS team_size,
       sum(s.amount) AS team_payroll,
       1 AS depth
FROM employees m
JOIN employees e ON e.manager_id = m.id
JOIN salaries s  ON s.employee_id = e.id
GROUP BY m.id, m.name
ORDER BY team_size DESC, manager;
