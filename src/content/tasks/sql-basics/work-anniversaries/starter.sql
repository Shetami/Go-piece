-- Годовщины работы с 2024-12-20 по 2025-03-05 включительно.
-- Колонки: name, hired_on, anniversary (date), years.
-- Порядок: anniversary, затем name.
SELECT name, hired_on,
       make_date(2024, extract(month FROM hired_on)::int, extract(day FROM hired_on)::int) AS anniversary,
       2024 - extract(year FROM hired_on)::int AS years
FROM employees
WHERE to_char(hired_on, 'MM-DD') BETWEEN '12-20' AND '12-31'
ORDER BY anniversary, name;
