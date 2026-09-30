CREATE TABLE products (
  sku   text PRIMARY KEY,
  name  text NOT NULL,
  price numeric(10, 2) NOT NULL
);

-- price пуст, если у конкурента товара нет в наличии и цену он не показывает
CREATE TABLE competitor_prices (
  competitor text NOT NULL,
  sku        text NOT NULL REFERENCES products (sku),
  price      numeric(10, 2),
  PRIMARY KEY (competitor, sku)
);

INSERT INTO products VALUES
  ('A-100', 'Чайник',      2490.00),
  ('A-200', 'Тостер',      3190.00),
  ('A-300', 'Блендер',     4990.00),
  ('A-400', 'Миксер',      2990.00),
  ('A-500', 'Кофемолка',   1990.00),
  ('A-600', 'Весы',         990.00);

INSERT INTO competitor_prices VALUES
  ('ozon',   'A-100', 2590.00),
  ('wb',     'A-100', 2790.00),
  ('ozon',   'A-200', 3190.00),
  ('wb',     'A-200', 3490.00),
  ('ozon',   'A-300', 5490.00),
  ('wb',     'A-300', NULL),
  ('ozon',   'A-400', 2890.00),
  ('wb',     'A-400', 3990.00),
  ('ozon',   'A-600', NULL),
  ('market', 'A-300', 5200.00);
