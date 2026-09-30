CREATE TABLE payments (
  id          int PRIMARY KEY,
  customer_id int  NOT NULL,
  plan        text NOT NULL,
  paid_at     date NOT NULL,
  amount      int,             -- рубли; у неуспешных платежей бывает NULL
  status      text NOT NULL CHECK (status IN ('paid', 'failed', 'refunded'))
);

-- refunded — отдельная строка возврата, сумма положительная.
INSERT INTO payments VALUES
  ( 1, 1, 'basic', '2024-01-03',  100, 'paid'),
  ( 2, 2, 'basic', '2024-01-05',  100, 'paid'),
  ( 3, 3, 'basic', '2024-01-09',  100, 'paid'),
  ( 4, 4, 'pro',   '2024-01-10',  300, 'paid'),
  ( 5, 4, 'pro',   '2024-01-20',  300, 'refunded'),
  ( 6, 5, 'team',  '2024-01-15', 1000, 'paid'),
  ( 7, 6, 'basic', '2024-01-31', NULL, 'failed'),
  ( 8, 1, 'basic', '2024-02-03',  100, 'paid'),
  ( 9, 4, 'pro',   '2024-02-10',  300, 'paid'),
  (10, 7, 'pro',   '2024-02-11',  300, 'paid'),
  (11, 5, 'team',  '2024-02-15', 1000, 'failed');
