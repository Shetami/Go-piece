CREATE TABLE orders (
  id         int PRIMARY KEY,
  created_on date NOT NULL,
  amount     int NOT NULL
);

INSERT INTO orders VALUES
  (1, '2024-03-01', 100),
  (2, '2024-03-01',  50),
  (3, '2024-03-02', 200),
  (4, '2024-03-03',  70),
  (5, '2024-03-03',  30),
  (6, '2024-03-04',  10);
