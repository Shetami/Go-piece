CREATE TABLE orders (
  id         int PRIMARY KEY,
  created_at timestamp NOT NULL,
  amount     int NOT NULL,
  status     text             -- NULL: заказ только создан, статуса ещё нет
);

INSERT INTO orders VALUES
  (1, '2024-12-23 10:00',    100, 'paid'),
  (2, '2024-12-29 23:59',    200, NULL),
  (3, '2024-12-30 00:00',    300, 'paid'),
  (4, '2024-12-31 18:00',     60, 'paid'),
  (5, '2025-01-01 12:00',    150, 'cancelled'),
  (6, '2025-01-01 13:00',     50, 'CANCELLED'),
  (7, '2025-01-05 22:00',    400, NULL),
  (8, '2025-01-06 00:00',    250, 'paid'),
  (9, '2025-01-12 23:59:59',  80, 'paid');
