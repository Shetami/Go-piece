-- Кто согласует заявку: руководитель, а если он отсутствует — руководитель руководителя.
-- Колонки: request_id, employee, approver (NULL, если некому), level (1, 2 или NULL).
-- Порядок: по request_id.
SELECT r.id AS request_id, e.name AS employee,
       coalesce(m1.name, m2.name) AS approver,
       CASE WHEN m1.name IS NOT NULL THEN 1 ELSE 2 END AS level
FROM expense_requests r
JOIN employees e  ON e.id = r.employee_id
JOIN employees m1 ON m1.id = e.manager_id
LEFT JOIN absences a ON a.employee_id = m1.id
JOIN employees m2 ON m2.id = m1.manager_id
WHERE a.employee_id IS NULL OR r.created_on NOT BETWEEN a.from_date AND a.to_date
ORDER BY r.id;
