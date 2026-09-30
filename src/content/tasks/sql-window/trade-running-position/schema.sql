CREATE TABLE trades (
  id      int PRIMARY KEY,
  account text      NOT NULL,
  ticker  text      NOT NULL,
  side    text      NOT NULL CHECK (side IN ('buy', 'sell')),
  qty     int       NOT NULL,
  ts      timestamp NOT NULL
);

-- Сделки 2 и 3 пришли в одну и ту же секунду.
-- У bob тоже есть AAPL — но это другой счёт.
INSERT INTO trades VALUES
  (1, 'alice', 'AAPL', 'buy',   10, '2024-03-01 09:00:00'),
  (4, 'bob',   'AAPL', 'buy',  100, '2024-03-01 09:01:00'),
  (5, 'alice', 'MSFT', 'buy',    3, '2024-03-01 09:02:00'),
  (2, 'alice', 'AAPL', 'buy',    5, '2024-03-01 09:05:00'),
  (3, 'alice', 'AAPL', 'sell',   8, '2024-03-01 09:05:00'),
  (7, 'bob',   'AAPL', 'sell',  40, '2024-03-01 09:10:00'),
  (6, 'alice', 'AAPL', 'sell',   7, '2024-03-01 09:30:00'),
  (8, 'alice', 'MSFT', 'sell',   3, '2024-03-01 10:00:00');
