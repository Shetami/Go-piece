WITH t AS (
  SELECT ticker, date_trunc('hour', ts) AS hour, price, qty,
         first_value(price) OVER w AS open,
         last_value(price)  OVER w AS close
  FROM trades
  WINDOW w AS (
    PARTITION BY ticker, date_trunc('hour', ts)
    ORDER BY ts, id
    ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING
  )
)
SELECT ticker,
       to_char(hour, 'YYYY-MM-DD HH24:MI') AS hour,
       open, max(price) AS high, min(price) AS low, close,
       sum(qty) AS volume
FROM t
GROUP BY ticker, t.hour, open, close
ORDER BY ticker, t.hour;
