CREATE TABLE deliveries (
  id      int PRIMARY KEY,
  courier text,              -- NULL: заказ так и не назначили
  status  text NOT NULL
);

INSERT INTO deliveries VALUES
  (1,  'Иван',   'done'),
  (2,  'Иван',   'cancelled'),
  (3,  'Иван',   'done'),
  (4,  'Иван',   'cancelled'),
  (5,  'Иван',   'done'),
  (6,  'Иван',   'cancelled'),
  (7,  'Олег',   'done'),
  (8,  'Олег',   'done'),
  (9,  'Олег',   'cancelled'),
  (10, 'Олег',   'done'),
  (11, 'Олег',   'done'),
  (12, 'Пётр',   'cancelled'),
  (13, 'Пётр',   'cancelled'),
  (14, 'Сергей', 'done'),
  (15, 'Сергей', 'done'),
  (16, 'Сергей', 'done'),
  (17, 'Сергей', 'done'),
  (18, 'Сергей', 'done'),
  (19, NULL,     'cancelled'),
  (20, NULL,     'cancelled'),
  (21, NULL,     'cancelled');
