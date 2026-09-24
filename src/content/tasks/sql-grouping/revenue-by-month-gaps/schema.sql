CREATE TABLE sales (
  id      int PRIMARY KEY,
  sold_at date NOT NULL,
  amount  int NOT NULL
);

INSERT INTO sales VALUES
  (1, '2024-01-05', 300),
  (2, '2024-01-20', 200),
  (3, '2024-02-11', 450),
  (4, '2024-04-01', 100),
  (5, '2024-04-30', 250),
  (6, '2024-06-15', 600),
  (7, '2024-07-02', 999);
