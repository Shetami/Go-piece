-- Баланс каждого клиента в рублях на конец дня 2024-03-31.
-- Колонки: client, accounts, balance_rub, place. Порядок: place, затем client.
SELECT c.name AS client,
       count(a.id) AS accounts,
       sum(t.amount * r.rub_rate) AS balance_rub,
       row_number() OVER (ORDER BY sum(t.amount * r.rub_rate) DESC) AS place
FROM clients c
JOIN accounts a     ON a.client_id = c.id
JOIN transactions t ON t.to_account = a.id
JOIN fx_rates r     ON r.currency = a.currency
WHERE t.ts <= '2024-03-31'
GROUP BY c.id, c.name
ORDER BY place, client;
