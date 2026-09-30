CREATE TABLE stores (
  id     int PRIMARY KEY,
  region text NOT NULL,
  city   text            -- NULL: онлайн-витрина региона
);

CREATE TABLE orders (
  id       int PRIMARY KEY,
  store_id int NOT NULL REFERENCES stores (id),
  amount   int NOT NULL,
  status   text NOT NULL
);

INSERT INTO stores VALUES
  (1, 'Центр', 'Москва'),
  (2, 'Центр', 'Москва'),
  (3, 'Центр', 'Тверь'),
  (4, 'Центр', NULL),
  (5, 'Урал',  'Екатеринбург'),
  (6, 'Урал',  'Пермь');

INSERT INTO orders VALUES
  (1,  1, 1000, 'paid'),
  (2,  1, 3000, 'paid'),
  (3,  2,  500, 'paid'),
  (4,  2,  700, 'refunded'),
  (5,  3, 1200, 'paid'),
  (6,  4,  400, 'paid'),
  (7,  4,  600, 'paid'),
  (8,  5, 2000, 'paid'),
  (9,  5, 1000, 'paid'),
  (10, 5, 3000, 'paid'),
  (11, 6,  900, 'refunded');
