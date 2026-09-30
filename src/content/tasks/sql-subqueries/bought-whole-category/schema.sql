CREATE TABLE products (
  id        int PRIMARY KEY,
  name      text NOT NULL,
  category  text NOT NULL,
  is_active boolean NOT NULL       -- false: снят с продажи
);

CREATE TABLE customers (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE purchases (
  id          int PRIMARY KEY,
  customer_id int NOT NULL REFERENCES customers (id),
  product_id  int NOT NULL REFERENCES products (id),
  returned    boolean NOT NULL DEFAULT false
);

INSERT INTO products VALUES
  (1, 'Эспрессо', 'Кофе',  true),
  (2, 'Арабика',  'Кофе',  true),
  (3, 'Робуста',  'Кофе',  true),
  (4, 'Мокко',    'Кофе',  false),
  (5, 'Пуэр',     'Чай',   true),
  (6, 'Улун',     'Чай',   true),
  (7, 'Цикорий',  'Архив', false);

INSERT INTO customers VALUES
  (1, 'Аня'), (2, 'Боря'), (3, 'Вера'), (4, 'Гоша'), (5, 'Даша'), (6, 'Егор');

INSERT INTO purchases (id, customer_id, product_id, returned) VALUES
  (1,  1, 1, false), (2,  1, 1, false), (3,  1, 2, false), (4,  1, 3, false),
  (5,  1, 5, false),
  (6,  2, 1, false), (7,  2, 1, false), (8,  2, 1, false), (9,  2, 2, false),
  (10, 3, 1, false), (11, 3, 2, false), (12, 3, 4, false),
  (13, 4, 1, false), (14, 4, 2, false), (15, 4, 3, true),
  (16, 5, 5, false), (17, 5, 6, false), (18, 5, 4, false),
  (19, 2, 7, false);
