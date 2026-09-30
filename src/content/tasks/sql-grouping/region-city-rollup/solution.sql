SELECT CASE WHEN GROUPING(s.region) = 1 THEN 'Все регионы' ELSE s.region END AS region,
       CASE WHEN GROUPING(s.city) = 1 THEN 'Итого' ELSE coalesce(s.city, 'онлайн') END AS city,
       count(o.id) FILTER (WHERE o.status = 'paid')                 AS orders,
       coalesce(sum(o.amount) FILTER (WHERE o.status = 'paid'), 0)  AS revenue,
       round(avg(o.amount) FILTER (WHERE o.status = 'paid'), 2)     AS avg_check
FROM stores s
LEFT JOIN orders o ON o.store_id = s.id
GROUP BY ROLLUP (s.region, s.city)
ORDER BY GROUPING(s.region), s.region, GROUPING(s.city), s.city NULLS LAST;
