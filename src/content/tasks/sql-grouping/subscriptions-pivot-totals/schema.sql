CREATE TABLE subscriptions (
  id         int PRIMARY KEY,
  user_id    int NOT NULL,
  plan       text NOT NULL,
  created_at timestamp NOT NULL,
  is_trial   boolean NOT NULL
);

INSERT INTO subscriptions VALUES
  (1,  1,  'basic',      '2024-01-05 10:00', false),
  (2,  2,  'pro',        '2024-01-10 12:00', false),
  (3,  3,  'basic',      '2024-01-20 09:00', true),
  (4,  4,  'team',       '2024-02-01 00:00', false),
  (5,  5,  'enterprise', '2024-02-15 15:00', false),
  (6,  6,  'pro',        '2024-02-29 23:59', false),
  (7,  7,  'pro',        '2024-03-10 11:00', true),
  (8,  8,  'basic',      '2024-04-01 00:00', false),
  (9,  9,  'basic',      '2024-04-30 12:00', false),
  (10, 10, 'pro',        '2024-05-01 00:00', false);
