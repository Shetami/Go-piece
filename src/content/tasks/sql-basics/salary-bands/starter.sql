-- Распределение сотрудников по вилкам оклада.
-- Колонки: band, employees, avg_salary (round(avg(salary))).
-- Порядок: низкая, средняя, высокая, не указана.
SELECT CASE WHEN salary < 100000 THEN 'низкая'
            WHEN salary < 200000 THEN 'средняя'
            ELSE 'высокая' END AS band,
       count(*) AS employees,
       round(avg(salary)) AS avg_salary
FROM employees
GROUP BY band
ORDER BY band;
