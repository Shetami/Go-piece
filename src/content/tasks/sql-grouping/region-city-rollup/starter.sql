-- Продажи по регионам и городам с подытогами.
-- Колонки: region ('Все регионы' в общем итоге), city (NULL-город -> 'онлайн',
--          подытог -> 'Итого'), orders, revenue, avg_check (round 2).
-- Порядок: регионы по алфавиту, внутри — города по алфавиту, 'онлайн', 'Итого';
--          строка 'Все регионы' — последней.
SELECT coalesce(s.region, 'Все регионы') AS region,
       coalesce(s.city, 'Итого') AS city,
       count(*) AS orders,
       sum(o.amount) AS revenue,
       round(avg(o.amount), 2) AS avg_check
FROM orders o
JOIN stores s ON s.id = o.store_id
WHERE o.status = 'paid'
GROUP BY ROLLUP (s.region, s.city)
ORDER BY 1, 2;
