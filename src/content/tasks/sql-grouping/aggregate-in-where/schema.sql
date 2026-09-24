CREATE TABLE orders (
  id       int PRIMARY KEY,
  customer text NOT NULL,
  amount   int NOT NULL
);

INSERT INTO orders VALUES
  (1, 'Аня',  500),
  (2, 'Аня',  300),
  (3, 'Боря', 900),
  (4, 'Аня',  200),
  (5, 'Вера', 100),
  (6, 'Боря', 400),
  (7, 'Вера', 150),
  (8, 'Вера', 250),
  (9, 'Вера',  50);
