-- Все платежи за январь 2024
SELECT id, amount, paid_at
FROM payments
WHERE paid_at BETWEEN '2024-01-01' AND '2024-01-31'
ORDER BY id;
