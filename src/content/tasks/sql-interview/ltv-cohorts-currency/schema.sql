CREATE TABLE users (
  id        int PRIMARY KEY,
  name      text NOT NULL,
  signup_on date NOT NULL,
  is_test   boolean NOT NULL DEFAULT false
);

-- Средний курс валюты к доллару за месяц. Доллара в таблице нет.
CREATE TABLE fx_monthly (
  month       date NOT NULL, -- первое число месяца
  currency    text NOT NULL,
  rate_to_usd numeric(10,4) NOT NULL,
  PRIMARY KEY (month, currency)
);

-- Заказ мог быть оформлен гостем и привязан к аккаунту позже регистрации
CREATE TABLE orders (
  id         int PRIMARY KEY,
  user_id    int NOT NULL REFERENCES users(id),
  created_on date NOT NULL,
  amount     numeric(12,2) NOT NULL,
  currency   text NOT NULL,
  status     text NOT NULL -- paid, refunded
);

INSERT INTO users VALUES
  (1, 'Аня',  '2024-01-10', false),
  (2, 'Боря', '2024-01-25', false),
  (3, 'Вера', '2024-01-31', false),
  (4, 'Гоша', '2024-02-05', false),
  (5, 'Дина', '2024-02-20', true),
  (6, 'Егор', '2024-02-29', false);

INSERT INTO fx_monthly VALUES
  ('2024-01-01', 'RUB', 0.0110), ('2024-02-01', 'RUB', 0.0108), ('2024-03-01', 'RUB', 0.0109),
  ('2024-01-01', 'EUR', 1.0900), ('2024-02-01', 'EUR', 1.0800), ('2024-03-01', 'EUR', 1.0850);

INSERT INTO orders VALUES
  (1, 1, '2024-01-10', 10000, 'RUB', 'paid'),
  (2, 1, '2024-02-08',  5000, 'RUB', 'paid'),
  (3, 1, '2024-02-09',   100, 'EUR', 'paid'),
  (4, 2, '2024-01-26',    50, 'USD', 'paid'),
  (5, 2, '2024-01-27',    20, 'USD', 'refunded'),
  (6, 3, '2024-01-30',    40, 'EUR', 'paid'),
  (7, 3, '2024-02-15',   100, 'EUR', 'paid'),
  (8, 5, '2024-02-21',  1000, 'EUR', 'paid'),
  (9, 6, '2024-03-01', 20000, 'RUB', 'paid');
