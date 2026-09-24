CREATE TABLE customers (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE orders (
  id          int PRIMARY KEY,
  customer_id int NOT NULL REFERENCES customers (id),
  created_at  date NOT NULL,
  amount      int NOT NULL
);

INSERT INTO customers VALUES
  (1, 'Аня'),
  (2, 'Боря'),
  (3, 'Вера'),
  (4, 'Гоша');

INSERT INTO orders VALUES
  (1, 1, '2023-11-20', 700),
  (2, 1, '2024-02-03', 1200),
  (3, 1, '2024-08-14', 300),
  (4, 2, '2023-06-01', 450),
  (5, 2, '2023-12-30', 900),
  (6, 4, '2024-04-22', 1500);
