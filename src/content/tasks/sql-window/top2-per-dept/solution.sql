SELECT dept, name, salary
FROM (
  SELECT dept, name, salary,
         row_number() OVER (PARTITION BY dept ORDER BY salary DESC, name) AS rn
  FROM employees
) ranked
WHERE rn <= 2
ORDER BY dept, salary DESC, name;
