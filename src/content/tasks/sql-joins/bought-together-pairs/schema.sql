CREATE TABLE products (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE orders (
  id     int PRIMARY KEY,
  status text NOT NULL  -- 'paid', 'cancelled'
);

-- Один товар может встретиться в заказе несколькими строками (разные размеры, акции).
-- product_id пуст у подарочной упаковки — это не товар.
CREATE TABLE order_items (
  order_id   int NOT NULL REFERENCES orders (id),
  product_id int REFERENCES products (id),
  qty        int NOT NULL
);

INSERT INTO products VALUES
  (1, 'Кофемолка'),
  (2, 'Кофе в зёрнах'),
  (3, 'Турка'),
  (4, 'Фильтры'),
  (5, 'Весы');

INSERT INTO orders VALUES
  (100, 'paid'),
  (101, 'paid'),
  (102, 'paid'),
  (103, 'cancelled'),
  (104, 'paid'),
  (105, 'paid'),
  (106, 'paid');

INSERT INTO order_items VALUES
  (100, 1, 1), (100, 2, 1), (100, 2, 2),
  (101, 1, 1), (101, 2, 1), (101, 4, 1), (101, NULL, 1),
  (102, 2, 3), (102, 3, 1),
  (103, 2, 1), (103, 3, 1),
  (104, 2, 1), (104, 4, 1), (104, NULL, 1),
  (105, 5, 1), (105, NULL, 1),
  (106, 2, 1), (106, 2, 1), (106, 3, 1), (106, 5, 1);
