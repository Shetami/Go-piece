SELECT to_char(created_at, 'IYYY-"W"IW')   AS week,         -- ISO-год, а не календарный
       date_trunc('week', created_at)::date AS week_start,  -- понедельник
       count(*)                             AS orders,
       sum(amount)                          AS revenue
FROM orders
WHERE lower(status) IS DISTINCT FROM 'cancelled'            -- NULL-статус не отмена
GROUP BY 1, 2
ORDER BY week_start;
