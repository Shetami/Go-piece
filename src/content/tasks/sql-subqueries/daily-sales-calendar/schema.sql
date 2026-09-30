CREATE TABLE orders (
  id         int PRIMARY KEY,
  created_at timestamp NOT NULL,
  amount     numeric(10, 2) NOT NULL,
  status     text NOT NULL             -- 'paid', 'cancelled'
);

INSERT INTO orders VALUES
  (1,  '2024-02-25 23:59:59', 700.00, 'paid'),
  (2,  '2024-02-26 00:00:00', 100.00, 'paid'),
  (3,  '2024-02-26 18:30:00', 250.00, 'paid'),
  (4,  '2024-02-27 12:00:00', 999.00, 'cancelled'),
  (5,  '2024-02-29 09:15:00', 400.00, 'paid'),
  (6,  '2024-02-29 23:59:59', 100.00, 'paid'),
  (7,  '2024-03-01 00:00:00',  50.00, 'paid'),
  (8,  '2024-03-03 23:59:59', 300.00, 'paid'),
  (9,  '2024-03-04 00:00:00', 800.00, 'paid');
