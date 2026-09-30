CREATE TABLE stock (
  warehouse  text NOT NULL,
  product_id int NOT NULL,
  qty        int NOT NULL,
  PRIMARY KEY (warehouse, product_id)
);

CREATE TABLE orders (
  id     int PRIMARY KEY,
  status text NOT NULL
);

-- Один товар может встречаться в заказе несколькими строками.
CREATE TABLE order_items (
  id         int PRIMARY KEY,
  order_id   int NOT NULL REFERENCES orders (id),
  product_id int NOT NULL,
  qty        int NOT NULL
);

INSERT INTO stock VALUES
  ('Север', 1, 5),
  ('Юг',    1, 3),
  ('Север', 2, 10),
  ('Юг',    3, 0);

INSERT INTO orders VALUES (1, 'new'), (2, 'new'), (3, 'new'), (4, 'shipped'), (5, 'new');

INSERT INTO order_items VALUES
  (1, 1, 1, 7),
  (2, 1, 2, 2),
  (3, 2, 1, 4),
  (4, 2, 1, 5),
  (5, 2, 2, 1),
  (6, 3, 4, 1),
  (7, 3, 2, 1),
  (8, 4, 1, 100),
  (9, 5, 3, 1);
