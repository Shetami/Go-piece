CREATE TABLE products (
  id  int PRIMARY KEY,
  sku text NOT NULL
);

CREATE TABLE receipts (
  id         int PRIMARY KEY,
  product_id int NOT NULL REFERENCES products (id),
  qty        int NOT NULL
);

CREATE TABLE orders (
  id     int PRIMARY KEY,
  status text NOT NULL
);

CREATE TABLE order_items (
  order_id   int NOT NULL REFERENCES orders (id),
  product_id int NOT NULL REFERENCES products (id),
  qty        int NOT NULL
);

CREATE TABLE returns (
  id         int PRIMARY KEY,
  product_id int NOT NULL REFERENCES products (id),
  qty        int NOT NULL
);

INSERT INTO products VALUES (1, 'A-100'), (2, 'B-200'), (3, 'C-300'), (4, 'D-400');

INSERT INTO receipts VALUES (1, 1, 50), (2, 1, 30), (3, 2, 10), (4, 3, 5);

INSERT INTO orders VALUES (1, 'paid'), (2, 'paid'), (3, 'cancelled'), (4, 'paid');

INSERT INTO order_items VALUES
  (1, 1, 10), (1, 2, 4),
  (2, 1, 5),
  (3, 1, 100), (3, 2, 7),
  (4, 2, 3), (4, 3, 6);

INSERT INTO returns VALUES (1, 1, 2), (2, 1, 1), (3, 2, 1);
