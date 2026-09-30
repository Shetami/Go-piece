-- Срок доставки и статус каждого отправления на 2024-05-13.
-- Колонки: id, due_date, delivered_on, status.
-- Порядок: due_date, затем id.
SELECT id,
       (accepted_at + sla_days * interval '1 day')::date AS due_date,
       delivered_at::date AS delivered_on,
       CASE WHEN delivered_at <= (accepted_at + sla_days * interval '1 day')::date THEN 'в срок'
            ELSE 'опоздание' END AS status
FROM shipments
ORDER BY due_date, id;
