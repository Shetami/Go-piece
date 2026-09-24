CREATE TABLE orders (
  id       int PRIMARY KEY,
  customer text NOT NULL,
  status   text NOT NULL,
  amount   int NOT NULL
);

INSERT INTO orders VALUES
  (1, 'Аня',  'paid',      1200),
  (2, 'Аня',  'cancelled',  300),
  (3, 'Аня',  'paid',       500),
  (4, 'Боря', 'new',        700),
  (5, 'Боря', 'cancelled',  250),
  (6, 'Вера', 'paid',        90),
  (7, 'Гоша', 'cancelled', 1000),
  (8, 'Гоша', 'cancelled',  400),
  (9, 'Аня',  'new',        150);
