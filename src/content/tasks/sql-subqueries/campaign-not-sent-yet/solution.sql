SELECT s.email
FROM subscribers s
WHERE s.unsubscribed_at IS NULL
  AND NOT EXISTS (
    SELECT 1
    FROM sends x
    WHERE x.campaign_id = 7
      AND x.subscriber_id = s.id
  )
ORDER BY s.email;
