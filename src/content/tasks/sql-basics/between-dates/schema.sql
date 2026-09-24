CREATE TABLE payments (
  id      int PRIMARY KEY,
  amount  int NOT NULL,
  paid_at timestamp NOT NULL
);

INSERT INTO payments VALUES
  (1, 500, '2023-12-31 23:59:00'),
  (2, 700, '2024-01-01 00:00:00'),
  (3, 300, '2024-01-15 10:20:00'),
  (4, 900, '2024-01-31 00:00:00'),
  (5, 450, '2024-01-31 18:45:00'),
  (6, 800, '2024-01-31 23:59:59'),
  (7, 100, '2024-02-01 00:00:00');
