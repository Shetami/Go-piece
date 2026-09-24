CREATE TABLE users (
  id         int PRIMARY KEY,
  name       text NOT NULL,
  last_login date
);

INSERT INTO users VALUES
  (1, 'Аня',  '2024-05-01'),
  (2, 'Боря', NULL),
  (3, 'Вера', '2024-06-10'),
  (4, 'Гоша', '2024-04-20'),
  (5, 'Даша', NULL);
