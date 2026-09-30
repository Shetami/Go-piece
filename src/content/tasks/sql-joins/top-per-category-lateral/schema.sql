CREATE TABLE categories (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE products (
  id          int PRIMARY KEY,
  name        text NOT NULL,
  category_id int NOT NULL REFERENCES categories (id)
);

CREATE TABLE orders (
  id         int PRIMARY KEY,
  created_at timestamp NOT NULL,
  status     text NOT NULL  -- 'delivered', 'cancelled'
);

CREATE TABLE order_items (
  order_id   int NOT NULL REFERENCES orders (id),
  product_id int NOT NULL REFERENCES products (id),
  qty        int NOT NULL
);

INSERT INTO categories VALUES
  (1, 'Зонты'),
  (2, 'Сумки'),
  (3, 'Шарфы'),
  (4, 'Перчатки');

INSERT INTO products VALUES
  (10, 'Зонт складной', 1),
  (11, 'Зонт-трость',   1),
  (12, 'Зонт детский',  1),
  (20, 'Рюкзак',        2),
  (21, 'Шопер',         2),
  (22, 'Клатч',         2),
  (30, 'Шарф шерстяной', 3),
  (40, 'Перчатки кожаные', 4);

INSERT INTO orders VALUES
  (1, '2024-05-02 10:00', 'delivered'),
  (2, '2024-05-10 12:00', 'delivered'),
  (3, '2024-05-15 18:00', 'cancelled'),
  (4, '2024-04-30 23:00', 'delivered'),
  (5, '2024-05-31 22:00', 'delivered'),
  (6, '2024-06-01 00:00', 'delivered');

INSERT INTO order_items VALUES
  (1, 10, 2),
  (1, 11, 1),
  (2, 11, 1),
  (2, 12, 2),
  (3, 12, 10),
  (1, 20, 1),
  (2, 21, 3),
  (5, 22, 3),
  (4, 20, 5),
  (3, 30, 4),
  (6, 40, 1),
  (5, 10, 1);
