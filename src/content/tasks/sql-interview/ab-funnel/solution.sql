WITH marked AS (
  SELECT e.user_id, e.ts, e.event, x.variant,
         CASE WHEN e.ts - lag(e.ts) OVER w <= interval '30 minutes' THEN 0 ELSE 1 END AS is_new
  FROM events e
  JOIN experiment x ON x.user_id = e.user_id
  JOIN users u      ON u.id = e.user_id
  WHERE NOT u.is_bot
  WINDOW w AS (PARTITION BY e.user_id ORDER BY e.ts)
),
sessions AS (
  SELECT *, sum(is_new) OVER (PARTITION BY user_id ORDER BY ts) AS session_no
  FROM marked
),
with_view AS (
  SELECT *, min(ts) FILTER (WHERE event = 'view') OVER s AS view_at
  FROM sessions
  WINDOW s AS (PARTITION BY user_id, session_no)
),
with_cart AS (
  SELECT *, min(ts) FILTER (WHERE event = 'cart' AND ts > view_at) OVER s AS cart_at
  FROM with_view
  WINDOW s AS (PARTITION BY user_id, session_no)
),
per_session AS (
  SELECT variant, user_id, session_no,
         bool_or(view_at IS NOT NULL)               AS viewed,
         bool_or(cart_at IS NOT NULL)               AS carted,
         bool_or(event = 'purchase' AND ts > cart_at) AS purchased
  FROM with_cart
  GROUP BY variant, user_id, session_no
)
SELECT variant,
       count(*)                            AS sessions,
       count(*) FILTER (WHERE viewed)      AS viewed,
       count(*) FILTER (WHERE carted)      AS carted,
       count(*) FILTER (WHERE purchased)   AS purchased,
       round(100.0 * count(*) FILTER (WHERE purchased)
                   / nullif(count(*) FILTER (WHERE viewed), 0), 1) AS conversion_pct
FROM per_session
GROUP BY variant
ORDER BY variant;
