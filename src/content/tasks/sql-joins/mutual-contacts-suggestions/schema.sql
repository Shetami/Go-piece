CREATE TABLE users (
  id         int PRIMARY KEY,
  name       text    NOT NULL,
  is_deleted boolean NOT NULL
);

-- Связь взаимная, но хранится одной строкой в произвольном направлении.
-- После миграции часть связей записана дважды — в обе стороны.
CREATE TABLE connections (
  user_a int NOT NULL REFERENCES users (id),
  user_b int NOT NULL REFERENCES users (id)
);

INSERT INTO users VALUES
  (1, 'Аня',   false),
  (2, 'Боря',  false),
  (3, 'Вера',  false),
  (4, 'Гоша',  false),
  (5, 'Дина',  false),
  (6, 'Егор',  true),
  (7, 'Жора',  false),
  (8, 'Зоя',   false),
  (9, 'Ира',   false);

INSERT INTO connections VALUES
  (1, 2),
  (3, 1),
  (2, 4),
  (4, 3),
  (3, 2),
  (2, 3),
  (5, 2),
  (1, 6),
  (6, 7),
  (4, 5),
  (8, 8),
  (9, 3),
  (2, 9);
