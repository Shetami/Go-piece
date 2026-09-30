CREATE TABLE customers (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE orders (
  id          int PRIMARY KEY,
  customer_id int NOT NULL REFERENCES customers (id),
  created_at  timestamp NOT NULL,
  amount      numeric(10, 2) NOT NULL,
  status      text NOT NULL            -- 'new', 'paid', 'cancelled'
);

INSERT INTO customers VALUES
  (1, 'Аня'),
  (2, 'Боря'),
  (3, 'Вера'),
  (4, 'Гоша'),
  (5, 'Даша');

INSERT INTO orders VALUES
  (1,  1, '2024-04-01 10:00', 1200.00, 'paid'),
  (2,  1, '2024-04-05 12:00',  300.00, 'paid'),
  (3,  1, '2024-04-07 09:00',  450.00, 'new'),
  (4,  1, '2024-04-08 18:00', 9900.00, 'cancelled'),
  (5,  2, '2024-04-02 11:00',  700.00, 'paid'),
  (6,  2, '2024-04-02 11:00',  150.00, 'paid'),
  (7,  2, '2024-04-02 11:00',  820.00, 'new'),
  (8,  3, '2024-04-03 15:30',   99.00, 'cancelled'),
  (9,  4, '2024-03-30 08:00', 5000.00, 'paid'),
  (10, 2, '2024-03-01 10:00', 2500.00, 'paid');
