CREATE TABLE orders (
  id     int PRIMARY KEY,
  status text NOT NULL
);

INSERT INTO orders VALUES
  (1, 'paid'),
  (2, 'cancelled'),
  (3, 'paid'),
  (4, 'shipped'),
  (5, 'cancelled'),
  (6, 'paid'),
  (7, 'cancelled'),
  (8, 'shipped');
