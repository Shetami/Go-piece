-- По два самых высокооплачиваемых сотрудника в каждом отделе.
-- При равной зарплате выше тот, чьё имя раньше по алфавиту. В отделе из одного человека — он один.
-- Колонки: dept, name, salary. Порядок: по dept, затем salary по убыванию, затем name.
SELECT dept, name, salary
FROM employees
ORDER BY dept, salary DESC, name
LIMIT 2;
