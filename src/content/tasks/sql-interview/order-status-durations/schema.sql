CREATE TABLE statuses (
  code       text PRIMARY KEY,
  sort_order int NOT NULL,
  is_final   boolean NOT NULL DEFAULT false
);

-- Журнал смены статусов. Вебхук иногда присылает тот же статус повторно.
CREATE TABLE status_log (
  order_id   int NOT NULL,
  status     text NOT NULL REFERENCES statuses(code),
  changed_at timestamp NOT NULL
);

INSERT INTO statuses VALUES
  ('created',    1, false),
  ('paid',       2, false),
  ('assembling', 3, false),
  ('on_hold',    4, false),
  ('shipped',    5, false),
  ('delivered',  6, true),
  ('cancelled',  7, true);

INSERT INTO status_log VALUES
  (1, 'created',    '2024-10-01 10:00'),
  (1, 'paid',       '2024-10-01 10:30'),
  (1, 'paid',       '2024-10-01 10:31'),
  (1, 'assembling', '2024-10-01 12:30'),
  (1, 'shipped',    '2024-10-02 12:30'),
  (1, 'delivered',  '2024-10-03 12:30'),

  (2, 'created',    '2024-10-01 09:00'),
  (2, 'paid',       '2024-10-01 10:00'),
  (2, 'assembling', '2024-10-01 11:00'),
  (2, 'paid',       '2024-10-01 12:00'),
  (2, 'assembling', '2024-10-01 14:00'),
  (2, 'shipped',    '2024-10-02 14:00'),

  (3, 'cancelled',  '2024-10-05 10:20'),
  (3, 'created',    '2024-10-05 10:00'),

  (4, 'created',    '2024-10-09 12:00');
