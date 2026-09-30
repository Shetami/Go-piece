-- Контакты вводились руками: регистр, пробелы, разный формат телефона,
-- пустые строки вместо NULL
CREATE TABLE customers (
  id    int PRIMARY KEY,
  name  text NOT NULL,
  email text,
  phone text
);

CREATE TABLE orders (
  id          int PRIMARY KEY,
  customer_id int NOT NULL REFERENCES customers(id),
  amount      numeric(10,2) NOT NULL
);

INSERT INTO customers VALUES
  (1,  'Анна',         'anna@mail.ru',     '+7 (900) 111-22-33'),
  (2,  'Анна К.',      ' ANNA@mail.ru ',   NULL),
  (3,  'Анна',         'anna.k@gmail.com', NULL),
  (4,  'Анна Козлова', 'kozlova@ya.ru',    '+7 900 111 22 33'),
  (5,  'Козлова',      'Kozlova@ya.ru',    ''),
  (6,  'Борис',        'boris@ya.ru',      ''),
  (7,  'Борис Л.',     '',                 ''),
  (8,  'Вера',         'vera@ya.ru',       '+7 999 000-00-00'),
  (9,  'Вера П.',      'vera.p@ya.ru',     '79990000000'),
  (10, 'Гоша',         '',                 NULL);

INSERT INTO orders VALUES
  (1, 1, 1000), (2, 2, 500), (3, 4, 700), (4, 5, 300), (5, 5, 200),
  (6, 3, 900),  (7, 6, 400), (8, 7, 100), (9, 8, 250), (10, 10, 50);
