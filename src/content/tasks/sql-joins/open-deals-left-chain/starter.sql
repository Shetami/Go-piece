-- Открытые сделки: клиент, менеджер, отдел и число звонков за неделю 24–30 июня 2024.
-- Колонки: deal_id, client, manager ('—', если нет), department ('—', если нет), calls_week.
-- Порядок: по deal_id.
SELECT d.id AS deal_id, d.client,
       coalesce(m.name, '—') AS manager,
       coalesce(dp.name, '—') AS department,
       count(*) AS calls_week
FROM deals d
LEFT JOIN managers m ON m.id = d.manager_id
JOIN departments dp  ON dp.id = m.department_id
LEFT JOIN calls c    ON c.deal_id = d.id
WHERE d.status = 'open'
  AND c.called_at >= '2024-06-24' AND c.called_at < '2024-07-01'
GROUP BY d.id, d.client, m.name, dp.name
ORDER BY d.id;
