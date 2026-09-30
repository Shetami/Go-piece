WITH inv AS (
  SELECT account_id, sum(amount) AS amount
  FROM invoices
  WHERE issued_on >= '2024-01-01' AND issued_on < '2025-01-01'
  GROUP BY account_id
),
pay AS (
  SELECT account_id, sum(amount) AS amount
  FROM payments
  WHERE paid_on >= '2024-01-01' AND paid_on < '2025-01-01'
  GROUP BY account_id
),
tix AS (
  SELECT account_id, count(*) AS open_tickets
  FROM tickets
  WHERE status = 'open'
  GROUP BY account_id
)
SELECT a.name AS account,
       c.email AS primary_contact,
       coalesce(i.amount, 0)                         AS invoiced,
       coalesce(p.amount, 0)                         AS paid,
       coalesce(i.amount, 0) - coalesce(p.amount, 0) AS balance,
       coalesce(t.open_tickets, 0)                   AS open_tickets
FROM accounts a
LEFT JOIN contacts c ON c.account_id = a.id AND c.is_primary
LEFT JOIN inv i ON i.account_id = a.id
LEFT JOIN pay p ON p.account_id = a.id
LEFT JOIN tix t ON t.account_id = a.id
WHERE a.is_active
ORDER BY balance DESC, a.name;
