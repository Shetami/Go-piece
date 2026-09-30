WITH b AS (
  SELECT salary,
         CASE WHEN salary IS NULL   THEN 4   -- NULL проверяем первым и явно
              WHEN salary < 100000 THEN 1
              WHEN salary < 200000 THEN 2
              ELSE 3 END AS k
  FROM employees
)
SELECT CASE k WHEN 1 THEN 'низкая' WHEN 2 THEN 'средняя'
              WHEN 3 THEN 'высокая' ELSE 'не указана' END AS band,
       count(*)            AS employees,
       round(avg(salary))  AS avg_salary
FROM b
GROUP BY k
ORDER BY k;
