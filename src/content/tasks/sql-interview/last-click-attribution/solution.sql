WITH attributed AS (
  SELECT o.id, o.amount,
         coalesce(lc.channel, 'direct') AS channel
  FROM orders o
  LEFT JOIN LATERAL (
    SELECT c.channel
    FROM touches t
    JOIN campaigns c ON c.id = t.campaign_id
    WHERE t.user_id = o.user_id
      AND t.kind = 'click'
      AND t.ts <= o.created_at
      AND t.ts >= o.created_at - interval '7 days'
    ORDER BY t.ts DESC, t.campaign_id
    LIMIT 1
  ) lc ON true
  WHERE o.status = 'paid'
),
channels AS (
  SELECT channel FROM campaigns
  UNION
  SELECT 'direct'
)
SELECT ch.channel,
       count(a.id)                 AS orders,
       coalesce(sum(a.amount), 0)  AS revenue,
       round(100.0 * coalesce(sum(a.amount), 0) / (SELECT sum(amount) FROM attributed), 1) AS share_pct
FROM channels ch
LEFT JOIN attributed a ON a.channel = ch.channel
GROUP BY ch.channel
ORDER BY revenue DESC, ch.channel;
