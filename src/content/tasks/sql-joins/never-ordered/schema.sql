CREATE TABLE products (
  id   int PRIMARY KEY,
  name text NOT NULL
);

-- product_id пуст у позиций «под заказ»: их собирают вручную, без карточки товара
CREATE TABLE order_items (
  order_id   int NOT NULL,
  product_id int REFERENCES products (id),
  qty        int NOT NULL
);

INSERT INTO products VALUES
  (1, 'Чайник'),
  (2, 'Кружка'),
  (3, 'Заварник'),
  (4, 'Ситечко'),
  (5, 'Поднос');

INSERT INTO order_items VALUES
  (100, 1,    1),
  (100, 2,    4),
  (101, 2,    2),
  (102, NULL, 1),
  (103, 5,    1);
