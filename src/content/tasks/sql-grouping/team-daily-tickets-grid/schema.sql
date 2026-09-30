CREATE TABLE teams (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE tickets (
  id         int PRIMARY KEY,
  team_id    int REFERENCES teams (id),   -- NULL: ещё не распределён
  created_at timestamp NOT NULL
);

INSERT INTO teams VALUES (1, 'Биллинг'), (2, 'Доставка'), (3, 'Возвраты');

INSERT INTO tickets VALUES
  (1,  1, '2024-06-30 23:00'),
  (2,  1, '2024-07-01 09:00'),
  (3,  1, '2024-07-01 23:59'),
  (4,  1, '2024-07-03 10:00'),
  (5,  1, '2024-07-03 11:00'),
  (6,  1, '2024-07-08 00:00'),
  (7,  2, '2024-07-02 12:00'),
  (8,  2, '2024-07-05 09:00'),
  (9,  2, '2024-07-05 10:00'),
  (10, 2, '2024-07-05 17:00'),
  (11, 2, '2024-07-07 20:00'),
  (12, NULL, '2024-07-04 10:00'),
  (13, NULL, '2024-07-04 11:00');
