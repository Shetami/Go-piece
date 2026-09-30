CREATE TABLE holidays (
  day date PRIMARY KEY
);

CREATE TABLE shipments (
  id           int PRIMARY KEY,
  accepted_at  timestamp NOT NULL,
  sla_days     int NOT NULL,      -- срок в рабочих днях
  delivered_at timestamp          -- NULL: ещё не доставлено
);

INSERT INTO holidays VALUES ('2024-05-01'), ('2024-05-09'), ('2024-05-10');

INSERT INTO shipments VALUES
  (1, '2024-04-26 10:00',    2, '2024-04-30 21:00'),
  (2, '2024-04-25 18:00',    2, '2024-04-30 10:00'),
  (3, '2024-04-30 12:00',    3, '2024-05-07 09:00'),
  (4, '2024-05-08 17:59:59', 2, NULL),
  (5, '2024-05-02 11:00',    1, NULL),
  (6, '2024-05-03 19:30',    1, '2024-05-06 23:59'),
  (7, '2024-05-07 09:00',    1, NULL),
  (8, '2024-05-06 10:00',    5, '2024-05-13 08:00');
