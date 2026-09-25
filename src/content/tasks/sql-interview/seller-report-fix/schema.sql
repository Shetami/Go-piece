CREATE TABLE sellers (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE products (
  id        int PRIMARY KEY,
  seller_id int NOT NULL REFERENCES sellers(id),
  title     text NOT NULL
);

CREATE TABLE orders (
  id         int PRIMARY KEY,
  created_at date NOT NULL,
  status     text NOT NULL -- delivered, cancelled
);

CREATE TABLE order_items (
  order_id   int NOT NULL REFERENCES orders(id),
  product_id int NOT NULL REFERENCES products(id),
  qty        int NOT NULL,
  price      int NOT NULL,
  PRIMARY KEY (order_id, product_id)
);

CREATE TABLE reviews (
  id         int PRIMARY KEY,
  product_id int NOT NULL REFERENCES products(id),
  rating     int NOT NULL CHECK (rating BETWEEN 1 AND 5)
);

INSERT INTO sellers VALUES
  (1, 'Лавка'),
  (2, 'Гаджетория'),
  (3, 'Книжный угол'),
  (4, 'Новичок');

INSERT INTO products VALUES
  (1, 1, 'Мёд'),
  (2, 1, 'Варенье'),
  (3, 2, 'Наушники'),
  (4, 2, 'Зарядка'),
  (5, 3, 'Атлас');

INSERT INTO orders VALUES
  (1, '2024-05-01', 'delivered'),
  (2, '2024-05-02', 'delivered'),
  (3, '2024-05-03', 'cancelled'),
  (4, '2024-05-04', 'delivered'),
  (5, '2024-05-05', 'delivered');

INSERT INTO order_items VALUES
  (1, 1, 2, 500),
  (1, 2, 1, 300),
  (2, 3, 1, 4000),
  (2, 4, 2, 1000),
  (3, 3, 1, 4000),
  (3, 5, 1, 1500),
  (4, 1, 1, 500),
  (5, 4, 1, 1000);

INSERT INTO reviews VALUES
  (1, 1, 5),
  (2, 1, 4),
  (3, 2, 3),
  (4, 3, 5),
  (5, 5, 4),
  (6, 5, 2);
