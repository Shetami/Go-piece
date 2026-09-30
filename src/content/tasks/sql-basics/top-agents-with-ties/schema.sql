CREATE TABLE agents (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE tickets (
  id        int PRIMARY KEY,
  agent_id  int NOT NULL REFERENCES agents(id),
  status    text NOT NULL,
  closed_at timestamp        -- NULL: тикет не закрыт
);

INSERT INTO agents VALUES
  (1, 'Анна'), (2, 'Борис'), (3, 'Вера'), (4, 'Гоша'), (5, 'Дина');

INSERT INTO tickets VALUES
  (1,  1, 'resolved',  '2024-03-01 00:00'),
  (2,  1, 'resolved',  '2024-03-05 10:00'),
  (3,  1, 'resolved',  '2024-03-10 11:00'),
  (4,  1, 'resolved',  '2024-03-15 12:00'),
  (5,  1, 'resolved',  '2024-03-20 13:00'),
  (6,  1, 'resolved',  '2024-02-29 23:59'),
  (7,  2, 'resolved',  '2024-03-02 09:00'),
  (8,  2, 'resolved',  '2024-03-03 09:00'),
  (9,  2, 'Resolved',  '2024-03-04 09:00'),
  (10, 2, 'resolved',  '2024-03-28 17:00'),
  (11, 2, 'open',      NULL),
  (12, 3, 'resolved',  '2024-03-06 14:00'),
  (13, 3, 'resolved',  '2024-03-07 15:00'),
  (14, 3, 'resolved',  '2024-03-08 16:00'),
  (15, 4, 'resolved',  '2024-03-09 10:00'),
  (16, 4, 'resolved',  '2024-03-12 10:00'),
  (17, 4, 'resolved',  '2024-03-31 23:30'),
  (18, 5, 'resolved',  '2024-03-11 10:00'),
  (19, 5, 'resolved',  '2024-03-13 10:00'),
  (20, 5, 'resolved',  '2024-04-01 00:00'),
  (21, 5, 'reopened',  '2024-03-14 10:00');
