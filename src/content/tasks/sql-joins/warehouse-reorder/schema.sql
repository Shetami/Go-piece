CREATE TABLE warehouses (
  id        int PRIMARY KEY,
  name      text    NOT NULL,
  is_active boolean NOT NULL
);

CREATE TABLE products (
  sku       text PRIMARY KEY,
  is_active boolean NOT NULL
);

-- Остаток по ячейкам: на один товар на складе бывает несколько строк.
-- Если строки нет — товара на складе нет.
CREATE TABLE stock (
  warehouse_id int  NOT NULL REFERENCES warehouses (id),
  sku          text NOT NULL REFERENCES products (sku),
  bin          text NOT NULL,
  qty          int  NOT NULL
);

-- Заказы поставщику: открытые ещё едут на склад
CREATE TABLE purchase_orders (
  id           int PRIMARY KEY,
  warehouse_id int  NOT NULL REFERENCES warehouses (id),
  sku          text NOT NULL REFERENCES products (sku),
  qty          int  NOT NULL,
  status       text NOT NULL  -- 'open', 'received'
);

-- Правила пополнения. warehouse_id пуст у общего правила для всех складов;
-- правило склада его заменяет. Правила редактируют, не удаляя старые строки:
-- действует последняя по updated_at.
CREATE TABLE reorder_rules (
  warehouse_id int REFERENCES warehouses (id),
  sku          text      NOT NULL REFERENCES products (sku),
  min_qty      int       NOT NULL,
  target_qty   int       NOT NULL,
  updated_at   timestamp NOT NULL
);

INSERT INTO warehouses VALUES
  (1, 'Москва', true),
  (2, 'Тверь',  true),
  (3, 'Клин',   false);

INSERT INTO products VALUES
  ('BOLT', true),
  ('NUT',  true),
  ('PIPE', true),
  ('OLD',  false);

INSERT INTO stock VALUES
  (1, 'BOLT', 'A-1', 30),
  (1, 'BOLT', 'A-2', 25),
  (1, 'NUT',  'B-1', 10),
  (2, 'BOLT', 'A-1', 70),
  (2, 'PIPE', 'C-1', 3),
  (1, 'OLD',  'Z-9', 0);

INSERT INTO purchase_orders VALUES
  (1, 1, 'NUT',  20, 'open'),
  (2, 1, 'NUT',  15, 'open'),
  (3, 2, 'PIPE', 40, 'received'),
  (4, 2, 'NUT',  50, 'open');

INSERT INTO reorder_rules VALUES
  (NULL, 'BOLT', 60,  100, '2024-01-01'),
  (2,    'BOLT', 50,  80,  '2024-01-01'),
  (2,    'BOLT', 80,  120, '2024-03-01'),
  (NULL, 'NUT',  40,  100, '2024-01-01'),
  (NULL, 'PIPE', 10,  20,  '2024-01-01'),
  (NULL, 'PIPE', 5,   15,  '2024-02-01'),
  (NULL, 'OLD',  10,  10,  '2024-01-01');
