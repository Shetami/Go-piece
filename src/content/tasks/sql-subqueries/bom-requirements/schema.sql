CREATE TABLE parts (
  id        int PRIMARY KEY,
  name      text NOT NULL,
  unit      text NOT NULL,
  unit_cost numeric(10, 2)       -- NULL у сборочных единиц: их не закупают
);

-- Спецификация: на одну единицу parent_id уходит qty единиц child_id
CREATE TABLE bom (
  parent_id int NOT NULL REFERENCES parts (id),
  child_id  int NOT NULL REFERENCES parts (id),
  qty       numeric(10, 3) NOT NULL,
  PRIMARY KEY (parent_id, child_id)
);

INSERT INTO parts VALUES
  (1,  'Велосипед',  'шт', NULL),
  (2,  'Рама',       'шт', NULL),
  (3,  'Колесо',     'шт', NULL),
  (4,  'Втулка',     'шт', NULL),
  (5,  'Труба',      'м',   400.00),
  (6,  'Болт',       'шт',   12.50),
  (7,  'Спица',      'шт',   15.00),
  (8,  'Обод',       'шт',  900.00),
  (9,  'Подшипник',  'шт',  150.00),
  (10, 'Трос',       'м',    80.00),
  (11, 'Самокат',    'шт', NULL),
  (12, 'Дека',       'шт', 1500.00);

INSERT INTO bom VALUES
  (1, 2,  1),
  (1, 3,  2),
  (1, 10, 1.5),
  (1, 6,  2),
  (2, 5,  2.4),
  (2, 6,  4),
  (3, 4,  1),
  (3, 7,  36),
  (3, 8,  1),
  (4, 9,  2),
  (4, 6,  1),
  (11, 12, 1),
  (11, 3,  2),
  (11, 6,  6);
