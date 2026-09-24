CREATE TABLE users (
  id         int PRIMARY KEY,
  name       text NOT NULL,
  email      text NOT NULL,
  deleted_at timestamp
);

INSERT INTO users VALUES
  (1, 'Аня',  'anya@example.com',  NULL),
  (2, 'Боря', 'borya@example.com', '2024-03-01 12:00'),
  (3, 'Вера', 'vera@example.com',  NULL),
  (4, 'Гоша', 'gosha@example.com', '2024-05-17 09:30'),
  (5, 'Даша', 'dasha@example.com', NULL);
