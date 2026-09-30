CREATE TABLE teams (
  id          int PRIMARY KEY,
  name        text NOT NULL,
  is_official boolean NOT NULL DEFAULT true
);

CREATE TABLE problems (
  id   int PRIMARY KEY,
  code text NOT NULL
);

CREATE TABLE contest (
  starts_at timestamp NOT NULL,
  ends_at   timestamp NOT NULL
);

-- verdict: OK, WA, TL, CE
CREATE TABLE submissions (
  id           int PRIMARY KEY,
  team_id      int NOT NULL REFERENCES teams(id),
  problem_id   int NOT NULL REFERENCES problems(id),
  submitted_at timestamp NOT NULL,
  verdict      text NOT NULL
);

INSERT INTO teams VALUES
  (1, 'Альфа', true), (2, 'Бета', true), (3, 'Гамма', true),
  (4, 'Дельта', true), (5, 'Гости', false), (6, 'Эпсилон', true);

INSERT INTO problems VALUES (1, 'A'), (2, 'B'), (3, 'C');

INSERT INTO contest VALUES ('2024-10-20 10:00', '2024-10-20 15:00');

INSERT INTO submissions VALUES
  (1,  1, 1, '2024-10-20 10:05', 'WA'),
  (2,  1, 1, '2024-10-20 10:30', 'OK'),
  (3,  1, 2, '2024-10-20 11:00', 'OK'),
  (4,  1, 2, '2024-10-20 11:10', 'WA'),
  (5,  1, 1, '2024-10-20 11:20', 'OK'),

  (6,  2, 1, '2024-10-20 10:10', 'CE'),
  (7,  2, 1, '2024-10-20 10:50:59', 'OK'),
  (8,  2, 2, '2024-10-20 11:00', 'OK'),

  (9,  3, 1, '2024-10-20 10:20', 'OK'),
  (10, 3, 2, '2024-10-20 10:30', 'WA'),
  (11, 3, 3, '2024-10-20 12:00', 'TL'),
  (12, 3, 3, '2024-10-20 15:00', 'OK'),

  (13, 4, 1, '2024-10-20 10:15', 'WA'),
  (14, 4, 1, '2024-10-20 10:25', 'WA'),
  (15, 4, 1, '2024-10-20 10:35', 'TL'),

  (16, 5, 1, '2024-10-20 10:01', 'OK'),
  (17, 5, 2, '2024-10-20 10:02', 'OK'),
  (18, 5, 3, '2024-10-20 10:03', 'OK');
