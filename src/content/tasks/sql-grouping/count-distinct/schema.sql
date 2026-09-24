CREATE TABLE orders (
  id       int PRIMARY KEY,
  customer text NOT NULL
);

CREATE TABLE order_items (
  order_id   int NOT NULL REFERENCES orders (id),
  product_id int NOT NULL,
  qty        int NOT NULL
);

INSERT INTO orders VALUES
  (1, 'Аня'),
  (2, 'Аня'),
  (3, 'Боря'),
  (4, 'Вера');

INSERT INTO order_items VALUES
  (1, 10, 1),
  (1, 20, 2),
  (2, 10, 1),
  (2, 30, 1),
  (3, 20, 5),
  (4, 40, 1),
  (4, 50, 1);
