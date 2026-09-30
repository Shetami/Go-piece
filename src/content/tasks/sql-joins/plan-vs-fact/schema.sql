-- План продаж: одна строка на регион и месяц
CREATE TABLE sales_plan (
  region     text    NOT NULL,
  month      date    NOT NULL,  -- первое число месяца
  amount     numeric NOT NULL,
  updated_by text    NOT NULL,
  PRIMARY KEY (region, month)
);

-- Факт: по каналам продаж, поэтому на регион и месяц бывает несколько строк.
-- region пуст у продаж, которые ещё не привязали к региону.
CREATE TABLE sales_fact (
  region     text,
  month      date    NOT NULL,
  channel    text    NOT NULL,
  amount     numeric NOT NULL,
  updated_by text    NOT NULL
);

INSERT INTO sales_plan VALUES
  ('Казань', '2023-12-01', 100, 'etl'),
  ('Казань', '2024-01-01', 120, 'etl'),
  ('Казань', '2024-02-01', 120, 'etl'),
  ('Москва', '2024-01-01', 500, 'etl'),
  ('Москва', '2024-02-01', 500, 'etl'),
  ('Москва', '2024-03-01', 600, 'etl'),
  ('Пермь',  '2024-02-01', 0,   'etl'),
  ('Пермь',  '2024-03-01', 80,  'etl');

INSERT INTO sales_fact VALUES
  ('Казань', '2023-12-01', 'shop',   90,  'etl'),
  ('Казань', '2024-01-01', 'shop',   70,  'etl'),
  ('Казань', '2024-01-01', 'online', 50,  'etl'),
  ('Москва', '2024-01-01', 'shop',   300, 'etl'),
  ('Москва', '2024-01-01', 'online', 250, 'etl'),
  ('Москва', '2024-02-01', 'shop',   500, 'etl'),
  ('Пермь',  '2024-02-01', 'shop',   40,  'etl'),
  ('Сочи',   '2024-03-01', 'shop',   60,  'etl'),
  ('Сочи',   '2024-04-01', 'shop',   70,  'etl'),
  (NULL,     '2024-02-01', 'online', 15,  'etl');
