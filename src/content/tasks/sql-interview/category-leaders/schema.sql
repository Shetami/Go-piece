-- Дерево категорий произвольной глубины: у корня parent_id IS NULL
CREATE TABLE categories (
  id        int PRIMARY KEY,
  name      text NOT NULL,
  parent_id int REFERENCES categories(id)
);

CREATE TABLE products (
  id          int PRIMARY KEY,
  name        text NOT NULL,
  category_id int NOT NULL REFERENCES categories(id),
  list_price  int NOT NULL -- цена в каталоге сейчас, а не цена продажи
);

CREATE TABLE customers (
  id      int PRIMARY KEY,
  name    text NOT NULL,
  is_test boolean NOT NULL DEFAULT false
);

CREATE TABLE orders (
  id          int PRIMARY KEY,
  customer_id int NOT NULL REFERENCES customers(id),
  created_at  timestamp NOT NULL,
  status      text NOT NULL -- pending, paid, shipped, cancelled
);

-- Цена — та, по которой товар продан в этом заказе
CREATE TABLE order_items (
  order_id   int NOT NULL REFERENCES orders(id),
  product_id int NOT NULL REFERENCES products(id),
  qty        int NOT NULL,
  price      int NOT NULL,
  PRIMARY KEY (order_id, product_id)
);

-- Одну позицию заказа могут возвращать частями, несколькими возвратами
CREATE TABLE refunds (
  id          int PRIMARY KEY,
  order_id    int NOT NULL,
  product_id  int NOT NULL,
  qty         int NOT NULL,
  refunded_at date NOT NULL,
  FOREIGN KEY (order_id, product_id) REFERENCES order_items(order_id, product_id)
);

INSERT INTO categories VALUES
  (1, 'Электроника', NULL),
  (2, 'Смартфоны',   1),
  (3, 'Ноутбуки',    1),
  (4, 'Аксессуары',  1),
  (5, 'Чехлы',       4),
  (6, 'Дом',         NULL),
  (7, 'Кухня',       6),
  (8, 'Текстиль',    6),
  (9, 'Книги',       NULL);

INSERT INTO products VALUES
  (1,  'Фон X',          2, 50000),
  (2,  'Фон Mini',       2, 30000),
  (3,  'Бук Pro',        3, 120000),
  (4,  'Бук Air',        3, 90000),
  (5,  'Кабель',         4, 1000),
  (6,  'Чехол кожаный',  5, 3000),
  (7,  'Чехол силикон',  5, 1000),
  (8,  'Сковорода',      7, 4000),
  (9,  'Чайник',         7, 3000),
  (10, 'Плед',           8, 5000),
  (11, 'Подушка',        8, 2000),
  (12, 'Роман',          9, 800),
  (13, 'Робот-пылесос',  6, 20000);

INSERT INTO customers VALUES
  (1, 'Аня',     false),
  (2, 'Боря',    false),
  (3, 'Вера',    false),
  (4, 'QA-бот',  true);

INSERT INTO orders VALUES
  (1,  1, '2024-01-10 12:00', 'paid'),
  (2,  2, '2024-01-15 09:00', 'shipped'),
  (3,  3, '2024-02-01 10:00', 'paid'),
  (4,  1, '2024-02-20 18:00', 'cancelled'),
  (5,  4, '2024-02-21 11:00', 'paid'),
  (6,  2, '2024-03-05 14:00', 'shipped'),
  (7,  3, '2024-03-31 23:30', 'paid'),
  (8,  1, '2024-04-01 00:10', 'paid'),
  (9,  2, '2023-12-31 23:59', 'paid'),
  (10, 3, '2024-03-15 10:00', 'pending');

INSERT INTO order_items VALUES
  (1,  1,  1, 48000), (1, 6, 2, 2500),  (1, 8, 1, 4000),
  (2,  3,  1, 115000), (2, 5, 3, 900),  (2, 10, 2, 5000),
  (3,  4,  1, 90000), (3, 2, 3, 29000), (3, 13, 1, 20000),
  (4,  3,  2, 120000),
  (5,  1,  5, 50000),
  (6,  1,  1, 47000), (6, 9, 5, 3000),  (6, 11, 3, 2000),
  (7,  4,  1, 88000), (7, 7, 4, 1000),  (7, 10, 1, 5000),
  (8,  3,  1, 120000),
  (9,  2,  3, 30000),
  (10, 13, 5, 20000);

INSERT INTO refunds VALUES
  (1, 2, 3,  1, '2024-02-01'),
  (2, 3, 2,  1, '2024-02-10'),
  (3, 3, 2,  1, '2024-04-15'),
  (4, 6, 11, 1, '2024-03-10');
