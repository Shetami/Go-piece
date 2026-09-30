-- Сколько сотрудников на каждом почтовом домене.
-- Колонки: domain (в нижнем регистре), employees.
-- Порядок: employees по убыванию, затем domain.
SELECT split_part(email, '@', 2) AS domain, count(*) AS employees
FROM employees
GROUP BY 1
ORDER BY employees DESC;
