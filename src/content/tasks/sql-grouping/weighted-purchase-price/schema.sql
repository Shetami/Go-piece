-- Строки закупок. qty = 0 — строка отменена поставщиком,
-- unit_price IS NULL — цена ещё не согласована.
CREATE TABLE purchases (
  id         int PRIMARY KEY,
  sku        text NOT NULL,
  qty        int NOT NULL,
  unit_price numeric(10, 2)
);

INSERT INTO purchases VALUES
  (1,  'A', 10,  100),
  (2,  'A', 90,  50),
  (3,  'A', 5,   NULL),
  (4,  'A', 0,   10),
  (5,  'B', 3,   200),
  (6,  'B', 4,   210),
  (7,  'B', 20,  NULL),
  (8,  'C', 6,   30),
  (9,  'C', 6,   40),
  (10, 'D', 10,  80);
