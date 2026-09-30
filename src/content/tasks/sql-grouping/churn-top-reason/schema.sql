CREATE TABLE plans (
  code text PRIMARY KEY
);

CREATE TABLE subscriptions (
  id        int PRIMARY KEY,
  plan_code text NOT NULL REFERENCES plans (code)
);

-- Пользователь может отменить подписку несколько раз (передумал, поменял причину):
-- в счёт идёт только последняя отмена. reason IS NULL — причину не указали.
CREATE TABLE cancellations (
  id              int PRIMARY KEY,
  subscription_id int NOT NULL REFERENCES subscriptions (id),
  cancelled_at    timestamp NOT NULL,
  reason          text
);

INSERT INTO plans VALUES ('basic'), ('pro'), ('team');

INSERT INTO subscriptions VALUES
  (1, 'basic'), (2, 'basic'), (3, 'basic'), (4, 'basic'), (5, 'basic'),
  (6, 'pro'), (7, 'pro'), (8, 'pro'), (9, 'pro'),
  (10, 'team');

INSERT INTO cancellations VALUES
  (1, 1, '2024-05-01 10:00', 'price'),
  (2, 1, '2024-05-03 10:00', 'bugs'),
  (3, 2, '2024-05-02 10:00', 'price'),
  (4, 3, '2024-05-04 10:00', NULL),
  (5, 4, '2024-05-05 10:00', 'bugs'),
  (6, 5, '2024-05-06 10:00', 'price'),
  (7, 6, '2024-05-01 10:00', 'switch'),
  (8, 7, '2024-05-02 10:00', 'price'),
  (9, 8, '2024-05-02 11:00', 'switch');
