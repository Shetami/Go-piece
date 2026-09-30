SELECT d.id AS deal_id,
       d.client,
       coalesce(m.name, '—')  AS manager,
       coalesce(dp.name, '—') AS department,
       count(c.id)            AS calls_week
FROM deals d
LEFT JOIN managers m     ON m.id = d.manager_id
LEFT JOIN departments dp ON dp.id = m.department_id
LEFT JOIN calls c
  ON c.deal_id = d.id
 AND c.called_at >= '2024-06-24' AND c.called_at < '2024-07-01'
WHERE d.status = 'open'
GROUP BY d.id, d.client, m.name, dp.name
ORDER BY d.id;
