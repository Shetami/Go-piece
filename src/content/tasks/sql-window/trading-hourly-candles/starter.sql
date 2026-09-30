-- Часовые свечи: open (цена первой сделки часа), high, low, close (цена последней), volume (сумма qty).
-- Порядок сделок: ts, при равенстве — id. Часы без сделок не выводятся.
-- Колонки: ticker, hour ('YYYY-MM-DD HH24:MI'), open, high, low, close, volume. Порядок: ticker, hour.
WITH t AS (
  SELECT ticker, date_trunc('hour', ts) AS hour, price, qty,
         first_value(price) OVER w AS open,
         last_value(price)  OVER w AS close
  FROM trades
  WINDOW w AS (PARTITION BY ticker, date_trunc('hour', ts) ORDER BY ts)
)
SELECT ticker, to_char(hour, 'YYYY-MM-DD HH24:MI') AS hour,
       min(open) AS open, max(price) AS high, min(price) AS low, max(close) AS close, sum(qty) AS volume
FROM t
GROUP BY ticker, t.hour
ORDER BY ticker, t.hour;
