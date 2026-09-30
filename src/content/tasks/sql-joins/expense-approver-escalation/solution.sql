WITH chain AS (
  SELECT r.id AS request_id, e.name AS employee, r.created_on,
         m1.id AS m1_id, m1.name AS m1_name,
         m2.id AS m2_id, m2.name AS m2_name
  FROM expense_requests r
  JOIN employees e       ON e.id = r.employee_id
  LEFT JOIN employees m1 ON m1.id = e.manager_id
  LEFT JOIN employees m2 ON m2.id = m1.manager_id
),
available AS (
  SELECT c.*,
         c.m1_id IS NOT NULL AND NOT EXISTS (
           SELECT 1 FROM absences a
           WHERE a.employee_id = c.m1_id
             AND c.created_on BETWEEN a.from_date AND a.to_date
         ) AS m1_ok,
         c.m2_id IS NOT NULL AND NOT EXISTS (
           SELECT 1 FROM absences a
           WHERE a.employee_id = c.m2_id
             AND c.created_on BETWEEN a.from_date AND a.to_date
         ) AS m2_ok
  FROM chain c
)
SELECT request_id, employee,
       CASE WHEN m1_ok THEN m1_name WHEN m2_ok THEN m2_name END AS approver,
       CASE WHEN m1_ok THEN 1 WHEN m2_ok THEN 2 END             AS level
FROM available
ORDER BY request_id;
