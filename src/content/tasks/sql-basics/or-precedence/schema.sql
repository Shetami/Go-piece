CREATE TABLE orders (
  id       int PRIMARY KEY,
  customer text NOT NULL,
  status   text NOT NULL,
  amount   int NOT NULL
);

INSERT INTO orders VALUES
  (1, 'Аня',  'paid',      1500),
  (2, 'Боря', 'shipped',    400),
  (3, 'Вера', 'new',       2500),
  (4, 'Гоша', 'shipped',   1200),
  (5, 'Даша', 'paid',       300),
  (6, 'Женя', 'shipped',     90),
  (7, 'Аня',  'cancelled', 5000);
