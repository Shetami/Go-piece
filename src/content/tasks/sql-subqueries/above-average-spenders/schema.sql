CREATE TABLE customers (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE orders (
  id          int PRIMARY KEY,
  customer_id int NOT NULL REFERENCES customers (id),
  amount      numeric(10, 2) NOT NULL,
  status      text NOT NULL            -- 'paid', 'cancelled'
);

INSERT INTO customers VALUES
  (1, 'Аня'), (2, 'Боря'), (3, 'Вера'), (4, 'Гоша'), (5, 'Даша'), (6, 'Егор');

INSERT INTO orders VALUES
  (1,  1, 100.00, 'paid'),
  (2,  1, 100.00, 'paid'),
  (3,  1, 100.00, 'paid'),
  (4,  1, 100.00, 'paid'),
  (5,  1, 100.00, 'paid'),
  (6,  1, 100.00, 'paid'),
  (7,  2, 700.00, 'paid'),
  (8,  3, 250.00, 'paid'),
  (9,  3, 250.00, 'paid'),
  (10, 4, 300.00, 'paid'),
  (11, 4, 900.00, 'cancelled'),
  (12, 5, 5000.00, 'cancelled');
