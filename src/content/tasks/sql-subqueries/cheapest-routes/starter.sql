-- Самый дешёвый маршрут из MOW в каждый достижимый город: не больше 3 перелётов,
-- ни один город не посещается дважды. Ничьи: меньше перелётов, затем route.
-- Колонки: city, legs, price, route ('MOW → LED → KZN'). Порядок: city.
SELECT DISTINCT ON (dst) dst AS city, 1 AS legs, price, src || ' → ' || dst AS route
FROM flights
WHERE src = 'MOW'
ORDER BY dst, price;
