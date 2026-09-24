SELECT shop, day, revenue,
       revenue - lag(revenue) OVER (ORDER BY shop, day) AS diff
FROM daily_revenue
ORDER BY shop, day;
