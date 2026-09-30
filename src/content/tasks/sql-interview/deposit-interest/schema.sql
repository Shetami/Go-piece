CREATE TABLE accounts (
  id     int PRIMARY KEY,
  client text NOT NULL
);

-- Годовая ставка действует с valid_from до следующей записи
CREATE TABLE rates (
  valid_from date PRIMARY KEY,
  rate_pct   numeric(6,2) NOT NULL
);

-- Пополнения (+) и списания (−); в один день бывает несколько операций
CREATE TABLE movements (
  account_id int NOT NULL REFERENCES accounts(id),
  op_date    date NOT NULL,
  amount     numeric(12,2) NOT NULL
);

INSERT INTO accounts VALUES (1, 'Аня'), (2, 'Боря'), (3, 'Вера'), (4, 'Гоша');

INSERT INTO rates VALUES
  ('2024-01-01', 14.64),
  ('2024-10-16', 18.30),
  ('2024-11-01', 20.00);

INSERT INTO movements VALUES
  (1, '2024-09-15', 100000),
  (1, '2024-10-10',  50000),
  (1, '2024-10-20', -30000),
  (2, '2024-10-25', 200000),
  (3, '2024-09-01',  10000),
  (3, '2024-10-05', -15000),
  (3, '2024-10-10',  12000),
  (3, '2024-10-10',   8000),
  (4, '2024-11-02',  50000);
