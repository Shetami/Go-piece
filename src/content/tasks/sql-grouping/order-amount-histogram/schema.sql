CREATE TABLE orders (
  id     int PRIMARY KEY,
  amount numeric(12, 2),
  status text NOT NULL
);

INSERT INTO orders VALUES
  (1,  150,     'paid'),
  (2,  999.99,  'paid'),
  (3,  1000,    'paid'),
  (4,  1500,    'paid'),
  (5,  1999,    'paid'),
  (6,  3000,    'paid'),
  (7,  3999.50, 'paid'),
  (8,  5000,    'paid'),
  (9,  12000,   'paid'),
  (10, 4500,    'cancelled'),
  (11, 2500,    'cancelled'),
  (12, NULL,    'paid');
