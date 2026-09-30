CREATE TABLE customers (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE products (
  id       int PRIMARY KEY,
  name     text NOT NULL,
  category text NOT NULL
);

-- product_id пуст, если товар удалили из каталога, а строку заказа оставили
CREATE TABLE order_items (
  id          int PRIMARY KEY,
  customer_id int NOT NULL REFERENCES customers (id),
  product_id  int REFERENCES products (id)
);

INSERT INTO customers VALUES
  (1, 'Аня'), (2, 'Боря'), (3, 'Вера'), (4, 'Гоша'), (5, 'Даша');

INSERT INTO products VALUES
  (1, 'Зерно Бразилия', 'Кофе'),
  (2, 'Зерно Кения',    'Кофе'),
  (3, 'Зерно Колумбия', 'Кофе'),
  (4, 'Турка',          'Посуда'),
  (5, 'Френч-пресс',    'Посуда'),
  (6, 'Пуэр',           'Чай');

INSERT INTO order_items VALUES
  (1,  1, 1),
  (2,  1, 1),
  (3,  1, 1),
  (4,  1, 1),
  (5,  2, 1),
  (6,  2, 2),
  (7,  3, 2),
  (8,  3, 4),
  (9,  3, 5),
  (10, 4, 6),
  (11, 4, NULL),
  (12, 5, 3),
  (13, 5, 4),
  (14, 2, NULL),
  (15, 4, 2),
  (16, 2, 4);
