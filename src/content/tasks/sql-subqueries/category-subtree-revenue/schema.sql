CREATE TABLE categories (
  id        int PRIMARY KEY,
  parent_id int REFERENCES categories (id),   -- NULL: корень каталога
  name      text NOT NULL
);

CREATE TABLE products (
  id          int PRIMARY KEY,
  category_id int NOT NULL REFERENCES categories (id),
  name        text NOT NULL
);

CREATE TABLE sales (
  id         int PRIMARY KEY,
  product_id int NOT NULL REFERENCES products (id),
  qty        int NOT NULL,
  price      numeric(10, 2) NOT NULL
);

INSERT INTO categories VALUES
  (1, NULL, 'Электроника'),
  (2, 1,    'Телефоны'),
  (3, 2,    'Смартфоны'),
  (4, 2,    'Кнопочные'),
  (5, 1,    'Ноутбуки'),
  (6, 3,    'Складные'),
  (7, NULL, 'Дом'),
  (8, 7,    'Кухня'),
  (9, 7,    'Текстиль');

INSERT INTO products VALUES
  (1, 3, 'Смартфон A'),
  (2, 3, 'Смартфон B'),
  (3, 6, 'Раскладушка Z'),
  (4, 2, 'Чехол универсальный'),
  (5, 5, 'Ноутбук X'),
  (6, 8, 'Чайник'),
  (7, 8, 'Сковорода'),
  (8, 4, 'Бабушкофон');

INSERT INTO sales VALUES
  (1,  1, 2, 30000.00),
  (2,  1, 1, 29000.00),
  (3,  2, 1, 45000.00),
  (4,  3, 1, 90000.00),
  (5,  4, 5,   500.00),
  (6,  5, 1, 80000.00),
  (7,  6, 3,  2000.00),
  (8,  6, 1,  1900.00);
