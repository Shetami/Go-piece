WITH RECURSIVE routes AS (
  SELECT dst AS city, 1 AS legs, price,
         ARRAY[src, dst] AS path
  FROM flights
  WHERE src = 'MOW'

  UNION ALL

  SELECT f.dst, r.legs + 1, r.price + f.price,
         r.path || f.dst
  FROM routes r
  JOIN flights f ON f.src = r.city
  WHERE r.legs < 3
    AND f.dst <> ALL (r.path)          -- не заходим в город второй раз
),
ranked AS (
  SELECT city, legs, price, array_to_string(path, ' → ') AS route,
         row_number() OVER (PARTITION BY city
                            ORDER BY price, legs, array_to_string(path, ' → ')) AS rn
  FROM routes
)
SELECT city, legs, price, route
FROM ranked
WHERE rn = 1
ORDER BY city;
