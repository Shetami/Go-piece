-- «Возможно, вы знакомы»: до двух кандидатов на пользователя по числу общих контактов.
-- Колонки: user_name, suggestion, mutual.
-- Порядок: по id пользователя, затем mutual по убыванию, затем по id кандидата.
SELECT u.name AS user_name, c.name AS suggestion, count(*) AS mutual
FROM connections f1
JOIN connections f2 ON f2.user_a = f1.user_b
JOIN users u ON u.id = f1.user_a
JOIN users c ON c.id = f2.user_b
WHERE f2.user_b <> f1.user_a
GROUP BY f1.user_a, f2.user_b, u.name, c.name
ORDER BY f1.user_a, mutual DESC, f2.user_b;
