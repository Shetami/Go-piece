CREATE TABLE customers (
  id   int PRIMARY KEY,
  name text NOT NULL
);

-- status пуст, если оплата ещё не подтверждена банком
CREATE TABLE orders (
  id          int PRIMARY KEY,
  customer_id int NOT NULL REFERENCES customers (id),
  status      text
);

INSERT INTO customers VALUES
  (1, 'Аня'),
  (2, 'Боря'),
  (3, 'Вера'),
  (4, 'Гоша'),
  (5, 'Даша');

INSERT INTO orders VALUES
  (1, 1, 'paid'),
  (2, 1, 'paid'),
  (3, 2, 'paid'),
  (4, 2, 'cancelled'),
  (5, 4, 'paid'),
  (6, 4, NULL),
  (7, 5, 'paid');
