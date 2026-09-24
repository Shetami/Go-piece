-- Рядом с каждой ценой — самая свежая цена того же тикера
SELECT ticker, day, price,
       last_value(price) OVER (
         PARTITION BY ticker
         ORDER BY day
       ) AS latest_price
FROM prices
ORDER BY ticker, day;
