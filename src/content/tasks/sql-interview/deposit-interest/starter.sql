-- Проценты по вкладам за октябрь 2024.
-- Колонки: client, avg_balance, interest. Порядок: interest по убыванию, затем client.
SELECT a.client,
       round(avg(m.amount), 2) AS avg_balance,
       round(sum(m.amount) * 14.64 / 100 / 12, 2) AS interest
FROM accounts a
JOIN movements m ON m.account_id = a.id
WHERE m.op_date BETWEEN '2024-10-01' AND '2024-10-31'
GROUP BY a.client
ORDER BY interest DESC, a.client;
