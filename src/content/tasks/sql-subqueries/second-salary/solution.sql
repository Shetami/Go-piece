SELECT max(salary) AS second_salary
FROM employees
WHERE salary < (SELECT max(salary) FROM employees);
