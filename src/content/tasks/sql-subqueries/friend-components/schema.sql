CREATE TABLE users (
  id   int PRIMARY KEY,
  name text NOT NULL
);

-- Дружба взаимна, но записана один раз в произвольном направлении
-- (а иногда — дважды, в обоих направлениях)
CREATE TABLE friendships (
  a int NOT NULL REFERENCES users (id),
  b int NOT NULL REFERENCES users (id),
  PRIMARY KEY (a, b)
);

INSERT INTO users VALUES
  (1, 'Аня'), (2, 'Боря'), (3, 'Вера'), (4, 'Гоша'), (5, 'Даша'),
  (6, 'Егор'), (7, 'Жора'), (8, 'Зина'), (9, 'Илья');

INSERT INTO friendships VALUES
  (1, 2), (2, 3), (3, 1), (2, 1),
  (4, 5), (6, 5), (9, 6),
  (7, 7);
