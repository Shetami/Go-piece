SELECT CASE WHEN GROUPING(d.day) = 1 THEN 'неделя'
            ELSE to_char(d.day, 'YYYY-MM-DD') END AS day,
       count(DISTINCT e.user_id) AS users,
       count(e.id)               AS events
FROM generate_series(timestamp '2024-10-07', timestamp '2024-10-13', interval '1 day') AS d(day)
LEFT JOIN events e
       ON e.happened_at >= d.day
      AND e.happened_at <  d.day + interval '1 day'
GROUP BY ROLLUP (d.day)
ORDER BY GROUPING(d.day), d.day;
