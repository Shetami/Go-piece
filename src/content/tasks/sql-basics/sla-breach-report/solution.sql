WITH t AS (
  SELECT coalesce(nullif(lower(trim(priority)), ''), 'normal') AS prio,
         -- без ответа ждём до момента отчёта
         coalesce(first_reply_at, timestamp '2024-06-10 12:00') - created_at AS waited
  FROM tickets
),
s AS (
  SELECT prio, waited,
         CASE prio WHEN 'high'   THEN interval '1 hour'
                   WHEN 'normal' THEN interval '4 hours'
                   WHEN 'low'    THEN interval '24 hours' END AS sla,
         CASE prio WHEN 'high' THEN 1 WHEN 'normal' THEN 2 ELSE 3 END AS sort_key
  FROM t
)
SELECT prio                                   AS priority,
       count(*)                               AS tickets,
       count(*) FILTER (WHERE waited > sla)   AS breached,
       round(100.0 * count(*) FILTER (WHERE waited > sla) / count(*), 1) AS breach_pct
FROM s
GROUP BY prio, sort_key
ORDER BY sort_key;
