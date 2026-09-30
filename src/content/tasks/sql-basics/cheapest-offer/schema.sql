CREATE TABLE products (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE sellers (
  id     int PRIMARY KEY,
  name   text NOT NULL,
  rating numeric(2,1)     -- NULL: новый продавец, оценок нет
);

CREATE TABLE offers (
  product_id int NOT NULL REFERENCES products(id),
  seller_id  int NOT NULL REFERENCES sellers(id),
  price      numeric(10,2) NOT NULL,
  stock      int,          -- NULL: склад не прислал остатки
  PRIMARY KEY (product_id, seller_id)
);

INSERT INTO products VALUES
  (1, 'Чайник'), (2, 'Термос'), (3, 'Турка'), (4, 'Кружка'), (5, 'Сито');

INSERT INTO sellers VALUES
  (1, 'Альфа', 4.8), (2, 'Бета', NULL), (3, 'Гамма', 4.8), (4, 'Дельта', 4.9);

INSERT INTO offers VALUES
  (1, 1, 1990.00, 5),
  (1, 2, 1790.00, NULL),
  (1, 3, 1890.00, 2),
  (1, 4, 1890.00, 0),
  (2, 1,  990.00, 3),
  (2, 2,  990.00, 10),
  (2, 3,  990.00, 1),
  (2, 4, 1100.00, 4),
  (3, 2,  550.00, 0),
  (4, 4,  300.00, 1),
  (4, 2,  300.00, 7);
