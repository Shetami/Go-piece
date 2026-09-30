CREATE TABLE customers (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE orders (
  id          int PRIMARY KEY,
  customer_id int NOT NULL REFERENCES customers (id),
  amount      numeric(10, 2) NOT NULL,
  status      text,                     -- NULL: заказ ещё не обработан
  created_at  timestamp NOT NULL
);

CREATE TABLE orders_archive (
  id          int PRIMARY KEY,
  customer_id int NOT NULL,
  amount      numeric(10, 2) NOT NULL,
  status      text,
  created_at  timestamp NOT NULL
);

INSERT INTO customers VALUES
  (1, 'Аня'), (2, 'Боря'), (3, 'Вера'), (4, 'Гоша');

INSERT INTO orders VALUES
  (1,  1, 100.00, 'cancelled', '2023-05-01 10:00'),
  (2,  1, 250.00, 'cancelled', '2023-12-31 23:59:59'),
  (3,  1, 900.00, 'paid',      '2023-06-01 12:00'),
  (4,  1,  70.00, 'cancelled', '2024-01-01 00:00'),
  (5,  2, 400.00, 'cancelled', '2023-02-14 09:00'),
  (6,  3, 300.00, NULL,        '2023-03-03 08:00'),
  (7,  3, 120.00, 'paid',      '2023-03-04 08:00'),
  (8,  4,  50.00, 'cancelled', '2023-08-08 18:00'),
  (9,  4,  60.00, 'paid',      '2024-02-02 11:00');

INSERT INTO orders_archive VALUES
  (100, 2, 999.00, 'cancelled', '2022-01-10 10:00'),
  (101, 2, 111.00, 'cancelled', '2022-05-10 10:00'),
  (102, 3,  10.00, 'cancelled', '2022-07-07 07:00');
