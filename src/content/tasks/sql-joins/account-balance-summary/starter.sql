-- Сводка по активным клиентам за 2024 год.
-- Колонки: account, primary_contact, invoiced, paid, balance (invoiced − paid), open_tickets.
-- Порядок: balance по убыванию, затем по account.
SELECT a.name AS account, c.email AS primary_contact,
       sum(i.amount) AS invoiced, sum(p.amount) AS paid,
       sum(i.amount) - sum(p.amount) AS balance,
       count(t.id) AS open_tickets
FROM accounts a
JOIN contacts c      ON c.account_id = a.id
LEFT JOIN invoices i ON i.account_id = a.id
LEFT JOIN payments p ON p.account_id = a.id
LEFT JOIN tickets t  ON t.account_id = a.id
WHERE a.is_active AND t.status = 'open'
GROUP BY a.name, c.email
ORDER BY balance DESC, account;
