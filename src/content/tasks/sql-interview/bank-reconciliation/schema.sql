-- Выписка банка: референс переписывается банком как попало
CREATE TABLE bank_statement (
  id        int PRIMARY KEY,
  tx_date   date NOT NULL,
  amount    numeric(10,2) NOT NULL,
  reference text
);

-- Платежи в нашей системе
CREATE TABLE payments (
  id        int PRIMARY KEY,
  paid_at   timestamp NOT NULL,
  amount    numeric(10,2) NOT NULL,
  reference text,
  status    text NOT NULL -- captured, failed
);

INSERT INTO payments VALUES
  (1, '2024-10-01 10:00', 100.00, 'INV-1', 'captured'),
  (2, '2024-10-01 11:00', 250.00, 'INV-2', 'captured'),
  (3, '2024-10-02 12:00', 500.00, NULL,    'captured'),
  (4, '2024-10-02 15:00', 500.00, NULL,    'captured'),
  (5, '2024-10-03 09:00',  75.00, 'INV-5', 'failed'),
  (6, '2024-10-03 10:00', 300.00, NULL,    'captured'),
  (7, '2024-10-04 10:00',  42.00, 'INV-7', 'captured'),
  (8, '2024-10-01 18:00', 120.00, NULL,    'captured');

INSERT INTO bank_statement VALUES
  (1, '2024-10-01', 100.00, ' inv-1 '),
  (2, '2024-10-02', 245.00, 'INV-2'),
  (3, '2024-10-03', 500.00, NULL),
  (4, '2024-10-03', 500.00, NULL),
  (5, '2024-10-03',  75.00, 'INV-5'),
  (6, '2024-10-06', 300.00, NULL),
  (7, '2024-10-01', 120.00, NULL),
  (8, '2024-10-01', 120.00, NULL);
