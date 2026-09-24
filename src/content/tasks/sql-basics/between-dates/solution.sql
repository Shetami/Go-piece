-- Все платежи за январь 2024
SELECT id, amount, paid_at
FROM payments
WHERE paid_at >= '2024-01-01' AND paid_at < '2024-02-01'
ORDER BY id;
