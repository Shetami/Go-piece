CREATE TABLE customers (
  id   int PRIMARY KEY,
  name text NOT NULL,
  city text NOT NULL
);

CREATE TABLE orders (
  id          int PRIMARY KEY,
  customer_id int NOT NULL REFERENCES customers (id),
  amount      int NOT NULL
);

INSERT INTO customers VALUES
  (1, 'Аня',  'Москва'),
  (2, 'Боря', 'Казань'),
  (3, 'Вера', 'Казань'),
  (4, 'Гоша', 'Пермь');

INSERT INTO orders VALUES
  (1, 1, 500),
  (2, 2, 300),
  (3, 3, 700),
  (4, 2, 150),
  (5, 4, 900);
