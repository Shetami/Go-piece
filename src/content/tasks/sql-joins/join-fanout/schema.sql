CREATE TABLE customers (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE orders (
  id          int PRIMARY KEY,
  customer_id int NOT NULL REFERENCES customers (id),
  amount      int NOT NULL
);

CREATE TABLE payments (
  id          int PRIMARY KEY,
  customer_id int NOT NULL REFERENCES customers (id),
  amount      int NOT NULL
);

INSERT INTO customers VALUES
  (1, 'Аня'),
  (2, 'Боря'),
  (3, 'Вера'),
  (4, 'Гоша');

INSERT INTO orders VALUES
  (1, 1, 1000),
  (2, 1,  500),
  (3, 2,  300),
  (4, 4,  200),
  (5, 4,  200),
  (6, 4,  100);

INSERT INTO payments VALUES
  (1, 1, 700),
  (2, 1, 800),
  (3, 3, 100),
  (4, 4, 500);
