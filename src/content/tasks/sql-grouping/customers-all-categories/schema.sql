CREATE TABLE categories (
  id        int PRIMARY KEY,
  name      text NOT NULL,
  is_active boolean NOT NULL
);

CREATE TABLE products (
  id          int PRIMARY KEY,
  category_id int REFERENCES categories (id)
);

CREATE TABLE orders (
  id       int PRIMARY KEY,
  customer text NOT NULL,
  status   text NOT NULL
);

CREATE TABLE order_items (
  id         int PRIMARY KEY,
  order_id   int NOT NULL REFERENCES orders (id),
  product_id int NOT NULL REFERENCES products (id)
);

INSERT INTO categories VALUES
  (1, 'Книги',  true),
  (2, 'Игры',   true),
  (3, 'Музыка', true),
  (4, 'Архив',  false);

INSERT INTO products VALUES (10, 1), (11, 1), (20, 2), (30, 3), (40, 4), (50, NULL);

INSERT INTO orders VALUES
  (1, 'Аня',  'paid'),
  (2, 'Аня',  'paid'),
  (3, 'Боря', 'paid'),
  (4, 'Боря', 'cancelled'),
  (5, 'Вера', 'paid'),
  (6, 'Гоша', 'paid'),
  (7, 'Гоша', 'paid');

INSERT INTO order_items VALUES
  (1,  1, 10), (2,  1, 11), (3, 1, 20),
  (4,  2, 30),
  (5,  3, 10), (6,  3, 20),
  (7,  4, 30),
  (8,  5, 10), (9,  5, 11), (10, 5, 20), (11, 5, 40), (12, 5, 50),
  (13, 6, 10), (14, 6, 20),
  (15, 7, 30), (16, 7, 30);
