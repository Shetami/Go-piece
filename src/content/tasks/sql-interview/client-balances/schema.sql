CREATE TABLE clients (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE accounts (
  id        int PRIMARY KEY,
  client_id int NOT NULL REFERENCES clients(id),
  currency  text NOT NULL -- RUB, USD, EUR
);

-- Пополнение: from_account IS NULL. Снятие: to_account IS NULL.
-- Перевод идёт только между счетами одной валюты; amount — в валюте счетов.
CREATE TABLE transactions (
  id           int PRIMARY KEY,
  ts           timestamp NOT NULL,
  from_account int REFERENCES accounts(id),
  to_account   int REFERENCES accounts(id),
  amount       numeric(14, 2) NOT NULL CHECK (amount > 0)
);

-- Курс публикуется не каждый день. Рубля в таблице нет.
CREATE TABLE fx_rates (
  currency  text NOT NULL,
  rate_date date NOT NULL,
  rub_rate  numeric(10, 4) NOT NULL, -- сколько рублей за единицу валюты
  PRIMARY KEY (currency, rate_date)
);

INSERT INTO clients VALUES
  (1, 'Аня'),
  (2, 'Боря'),
  (3, 'Вера'),
  (4, 'Гоша'),
  (5, 'Дина');

INSERT INTO accounts VALUES
  (1, 1, 'RUB'),
  (2, 1, 'USD'),
  (3, 2, 'EUR'),
  (4, 3, 'RUB'),
  (5, 3, 'USD'),
  (6, 4, 'RUB');

INSERT INTO transactions VALUES
  (1, '2024-01-10 10:00', NULL, 1,    100000),
  (2, '2024-02-01 10:00', NULL, 2,    1000),
  (3, '2024-02-15 12:00', 1,    4,    30000),
  (4, '2024-03-05 09:00', 4,    NULL, 5000),
  (5, '2024-03-10 15:00', 2,    5,    200),
  (6, '2024-03-20 11:00', NULL, 3,    627.50),
  (7, '2024-03-25 16:00', 3,    NULL, 100),
  (8, '2024-03-31 23:59', NULL, 5,    100),
  (9, '2024-04-01 00:01', NULL, 1,    50000);

INSERT INTO fx_rates VALUES
  ('USD', '2024-03-01', 90.0000),
  ('USD', '2024-03-29', 92.5000),
  ('USD', '2024-04-01', 95.0000),
  ('EUR', '2024-02-01', 98.0000),
  ('EUR', '2024-03-15', 100.0000);
