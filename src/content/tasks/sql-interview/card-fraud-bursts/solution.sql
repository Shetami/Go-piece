WITH p AS (
  SELECT pm.*
  FROM payments pm
  JOIN users u ON u.id = pm.user_id
  WHERE NOT u.is_test
),
windows AS (
  -- окно начинается в момент каждой оплаты и длится 10 минут включительно
  SELECT a.id, a.user_id, a.created_at AS started_at,
         count(DISTINCT b.card_id)                        AS cards,
         count(*) FILTER (WHERE b.status = 'declined')    AS declined
  FROM p a
  JOIN p b ON b.user_id = a.user_id
          AND b.created_at >= a.created_at
          AND b.created_at <= a.created_at + interval '10 minutes'
  GROUP BY a.id, a.user_id, a.created_at
),
ranked AS (
  SELECT w.*,
         max(cards) OVER (PARTITION BY user_id) AS max_cards,
         row_number() OVER (PARTITION BY user_id ORDER BY started_at, id) AS rn
  FROM windows w
  WHERE cards >= 3
)
SELECT u.name AS user_name,
       r.started_at AS burst_start,
       r.cards AS cards_in_burst,
       r.declined AS declined_in_burst,
       r.max_cards
FROM ranked r
JOIN users u ON u.id = r.user_id
WHERE r.rn = 1
ORDER BY burst_start, user_name;
