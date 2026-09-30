CREATE TABLE products (
  id       int PRIMARY KEY,
  name     text NOT NULL,
  category text
);

CREATE TABLE sales (
  id         int PRIMARY KEY,
  product_id int NOT NULL REFERENCES products (id),
  amount     numeric(12, 2) NOT NULL,
  currency   text NOT NULL,
  status     text NOT NULL
);

-- Курс рубля к рублю не хранится.
CREATE TABLE rates (
  currency    text PRIMARY KEY,
  rate_to_rub numeric(10, 4) NOT NULL
);

INSERT INTO rates VALUES ('USD', 90), ('EUR', 100);

INSERT INTO products VALUES
  (1, 'Ноутбук',          'Электроника'),
  (2, 'Наушники',         'Электроника'),
  (3, 'Чайник',           'Кухня'),
  (4, 'Сковорода',        'Кухня'),
  (5, 'Подарочная карта', NULL);

INSERT INTO sales VALUES
  (1, 1, 1000, 'USD', 'completed'),
  (2, 2,   50, 'usd', 'completed'),
  (3, 3, 3000, 'RUB', 'completed'),
  (4, 4,   20, 'EUR', 'completed'),
  (5, 4,   25, 'EUR', 'cancelled'),
  (6, 5, 5000, 'RUB', 'completed'),
  (7, 1,  900, 'USD', 'refunded'),
  (8, 3, 2500, 'rub', 'completed');
