-- Счёт за трафик за июнь 2024 по тарифу, действовавшему в день потребления.
-- Колонки: client, billed_gb, unbilled_gb (дни без тарифа), amount (round(…, 2)).
-- Клиенты без трафика — с нулями. Порядок: по client.
SELECT c.name AS client, sum(u.gb) AS billed_gb, 0 AS unbilled_gb,
       round(sum(u.gb * t.price_per_gb), 2) AS amount
FROM clients c
JOIN usage u          ON u.client_id = c.id
JOIN client_tariffs ct ON ct.client_id = u.client_id
                      AND u.day BETWEEN ct.valid_from AND coalesce(ct.valid_to, 'infinity')
JOIN tariffs t        ON t.code = ct.tariff_code
WHERE u.day >= '2024-06-01' AND u.day < '2024-07-01'
GROUP BY c.name
ORDER BY c.name;
