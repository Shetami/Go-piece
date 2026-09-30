SELECT id, account, ticker,
       sum(CASE side WHEN 'buy' THEN qty ELSE -qty END) OVER (
         PARTITION BY account, ticker
         ORDER BY ts, id
         ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
       ) AS position
FROM trades
ORDER BY account, ticker, ts, id;
