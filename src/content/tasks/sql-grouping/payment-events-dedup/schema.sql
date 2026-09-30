-- Вебхуки платёжного провайдера. Доставка at-least-once: одно и то же
-- событие (event_id) может прийти несколько раз.
CREATE TABLE payment_events (
  event_id    text NOT NULL,
  payment_id  int NOT NULL,
  status      text NOT NULL,
  amount      int NOT NULL,
  received_at timestamp NOT NULL
);

INSERT INTO payment_events VALUES
  ('e1', 1, 'succeeded', 1000, '2024-09-01 10:00'),
  ('e1', 1, 'succeeded', 1000, '2024-09-01 10:00'),
  ('e2', 2, 'succeeded',  500, '2024-09-01 11:00'),
  ('e3', 3, 'failed',     700, '2024-09-01 12:00'),
  ('e4', 2, 'refunded',   500, '2024-09-02 09:00'),
  ('e4', 2, 'refunded',   500, '2024-09-02 09:00'),
  ('e5', 4, 'succeeded',  300, '2024-09-02 10:00'),
  ('e6', 5, 'pending',    200, '2024-09-03 08:00'),
  ('e7', 1, 'refunded',   400, '2024-09-03 12:00');
