-- Пользователи, сменившие 3+ карты за 10 минут.
-- Колонки: user_name, burst_start, cards_in_burst, declined_in_burst, max_cards.
-- Порядок: burst_start, затем user_name.
SELECT u.name AS user_name,
       min(p.created_at) AS burst_start,
       count(p.card_id) AS cards_in_burst,
       count(*) FILTER (WHERE p.status = 'declined') AS declined_in_burst,
       count(p.card_id) AS max_cards
FROM payments p
JOIN users u ON u.id = p.user_id
GROUP BY u.name, date_trunc('hour', p.created_at)
HAVING count(DISTINCT p.card_id) >= 3
ORDER BY burst_start, user_name;
