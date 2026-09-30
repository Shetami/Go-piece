CREATE TABLE customers (
  id   int PRIMARY KEY,
  name text NOT NULL,
  city text NOT NULL
);

CREATE TABLE purchases (
  id          int PRIMARY KEY,
  customer_id int NOT NULL REFERENCES customers (id),
  channel     text NOT NULL,         -- 'online', 'store'
  amount      numeric(10, 2) NOT NULL
);

INSERT INTO customers VALUES
  (1, 'Аня',  'Казань'),
  (2, 'Боря', 'Казань'),
  (3, 'Вера', 'Казань'),
  (4, 'Гоша', 'Москва'),
  (5, 'Даша', 'Москва'),
  (6, 'Егор', 'Пермь'),
  (7, 'Жора', 'Пермь'),
  (8, 'Зина', 'Томск');

INSERT INTO purchases VALUES
  (1,  1, 'online', 500.00),
  (2,  1, 'online', 700.00),
  (3,  1, 'store',  300.00),
  (4,  1, 'store',  100.00),
  (5,  2, 'online', 900.00),
  (6,  3, 'store',  200.00),
  (7,  3, 'online', 150.00),
  (8,  4, 'store',  400.00),
  (9,  4, 'store',  450.00),
  (10, 5, 'online', 800.00),
  (11, 5, 'store',  100.00),
  (12, 5, 'online', 200.00),
  (13, 6, 'online', 300.00);
