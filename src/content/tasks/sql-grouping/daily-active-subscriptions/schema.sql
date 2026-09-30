-- Подписка действует в дни d, для которых started_at <= d < ended_at.
-- ended_at IS NULL — подписка не закончилась.
CREATE TABLE subscriptions (
  id         int PRIMARY KEY,
  user_id    int NOT NULL,
  plan       text NOT NULL,
  started_at date NOT NULL,
  ended_at   date
);

INSERT INTO subscriptions VALUES
  (1, 1, 'basic', '2024-05-20', '2024-06-03'),
  (2, 1, 'pro',   '2024-06-02', '2024-06-05'),
  (3, 2, 'pro',   '2024-06-01', '2024-06-01'),
  (4, 3, 'basic', '2024-05-01', '2024-06-01'),
  (5, 4, 'basic', '2024-06-04', '2024-06-05'),
  (6, 5, 'pro',   '2024-06-07', NULL),
  (7, 6, 'basic', '2024-06-08', NULL),
  (8, 2, 'basic', '2024-05-30', '2024-06-02');
