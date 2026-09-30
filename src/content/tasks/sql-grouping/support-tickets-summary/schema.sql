CREATE TABLE tickets (
  id          int PRIMARY KEY,
  priority    text,
  created_at  timestamp NOT NULL,
  resolved_at timestamp,
  csat        int
);

INSERT INTO tickets VALUES
  (1, 'high',   '2024-03-01 09:00', '2024-03-01 11:00', 5),
  (2, 'high',   '2024-03-01 10:00', '2024-03-02 10:00', NULL),
  (3, 'high',   '2024-03-02 12:00', NULL,               NULL),
  (4, 'low',    '2024-03-01 08:00', '2024-03-03 08:00', 3),
  (5, 'low',    '2024-03-04 09:00', NULL,               NULL),
  (6, NULL,     '2024-03-02 09:00', '2024-03-02 09:30', 4),
  (7, NULL,     '2024-03-05 10:00', NULL,               NULL),
  (8, 'normal', '2024-03-03 10:00', NULL,               NULL),
  (9, 'normal', '2024-03-03 11:00', NULL,               NULL);
