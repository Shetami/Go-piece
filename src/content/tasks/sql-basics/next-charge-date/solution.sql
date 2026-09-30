SELECT s.id, s.customer,
       (SELECT min(d)
        FROM (SELECT (s.started_on + n * interval '1 month')::date AS d
              FROM generate_series(0, 120) AS n) AS charges       -- от якоря, а не от прошлого списания
        WHERE d >= date '2024-03-30'                                -- сегодняшнее списание ещё впереди
          AND (s.canceled_on IS NULL OR d < s.canceled_on)          -- в день отмены уже не списываем
       ) AS next_charge
FROM subscriptions s
ORDER BY next_charge NULLS LAST, s.id;
