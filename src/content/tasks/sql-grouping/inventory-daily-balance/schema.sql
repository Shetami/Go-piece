CREATE TABLE products (
  sku text PRIMARY KEY
);

-- qty > 0 — приход, qty < 0 — расход.
CREATE TABLE stock_moves (
  id       int PRIMARY KEY,
  sku      text NOT NULL REFERENCES products (sku),
  moved_at timestamp NOT NULL,
  qty      int NOT NULL
);

INSERT INTO products VALUES ('BOLT'), ('NUT'), ('WASHER');

INSERT INTO stock_moves VALUES
  (1, 'BOLT', '2024-07-20 12:00',  100),
  (2, 'BOLT', '2024-08-01 10:00',  -30),
  (3, 'BOLT', '2024-08-01 15:00',   10),
  (4, 'BOLT', '2024-08-03 23:59',  -50),
  (5, 'BOLT', '2024-08-06 00:00',  -20),
  (6, 'NUT',  '2024-07-31 23:59',    5),
  (7, 'NUT',  '2024-08-04 08:00',   20);
