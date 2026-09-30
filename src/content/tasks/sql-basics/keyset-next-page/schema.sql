CREATE TABLE products (
  id        int PRIMARY KEY,
  name      text NOT NULL,
  rating    numeric(2,1),          -- NULL: оценок ещё нет
  is_active boolean NOT NULL
);

INSERT INTO products VALUES
  (1,  'Чайник',       4.8, true),
  (2,  'Кружка',       NULL, true),
  (3,  'Френч-пресс',  4.5, false),
  (4,  'Турка',        4.9, true),
  (5,  'Кофемолка',    3.9, true),
  (6,  'Сахарница',    4.0, false),
  (7,  'Термос',       4.2, true),
  (8,  'Заварник',     4.7, true),
  (9,  'Поднос',       NULL, true),
  (10, 'Сито',         4.5, true),
  (11, 'Ложка',        NULL, false),
  (12, 'Молочник',     4.5, true),
  (13, 'Фильтры',      5.0, true),
  (14, 'Весы',         NULL, true),
  (15, 'Капсулы',      4.5, true),
  (18, 'Банка',        4.5, false),
  (20, 'Кофе в зёрнах', 4.5, true);
