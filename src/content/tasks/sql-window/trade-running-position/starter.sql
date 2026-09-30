-- Позиция (сколько бумаг на руках) после каждой сделки — по счёту и тикеру.
-- buy добавляет qty, sell вычитает. Сделки в одну секунду — по возрастанию id.
-- Колонки: id, account, ticker, position. Порядок: account, ticker, ts, id.
SELECT id, account, ticker,
       sum(qty) OVER (PARTITION BY ticker ORDER BY ts) AS position
FROM trades
ORDER BY account, ticker, ts, id;
