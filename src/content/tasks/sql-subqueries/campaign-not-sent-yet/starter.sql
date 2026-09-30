-- Активные подписчики, которым кампанию 7 ещё не отправляли.
-- Колонка: email. Порядок: по email.
SELECT s.email
FROM subscribers s
WHERE s.unsubscribed_at IS NULL
  AND s.id NOT IN (SELECT subscriber_id FROM sends WHERE campaign_id = 7)
ORDER BY s.email;
