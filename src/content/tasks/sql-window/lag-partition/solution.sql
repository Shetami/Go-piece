SELECT shop, day, revenue,
       revenue - lag(revenue) OVER (PARTITION BY shop ORDER BY day) AS diff
FROM daily_revenue
ORDER BY shop, day;
