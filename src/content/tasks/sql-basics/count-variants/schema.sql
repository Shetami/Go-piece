CREATE TABLE customers (
  id    int PRIMARY KEY,
  name  text NOT NULL,
  phone text,
  city  text
);

INSERT INTO customers VALUES
  (1, 'Аня',  '+7 900 111-11-11', 'Москва'),
  (2, 'Боря', NULL,               'Казань'),
  (3, 'Вера', '+7 900 333-33-33', NULL),
  (4, 'Гоша', '+7 900 444-44-44', 'Москва'),
  (5, 'Даша', NULL,               'Казань'),
  (6, 'Женя', '+7 900 666-66-66', NULL);
