-- Вебхуки платёжного провайдера. Провайдер шлёт событие на каждую смену
-- статуса и при таймауте повторяет доставку — одинаковые события приходят дважды.
CREATE TABLE provider_events (
  event_id    int PRIMARY KEY,
  external_id text      NOT NULL,
  status      text      NOT NULL,  -- 'pending', 'succeeded', 'failed', 'refunded'
  amount      numeric   NOT NULL,
  received_at timestamp NOT NULL
);

-- Внутренний журнал поступлений. external_id пуст у ручных проводок бухгалтерии.
CREATE TABLE ledger (
  id          int PRIMARY KEY,
  external_id text,
  amount      numeric   NOT NULL,
  created_at  timestamp NOT NULL
);

INSERT INTO provider_events VALUES
  (1,  'pay_A', 'pending',   100, '2024-07-01 10:00'),
  (2,  'pay_A', 'succeeded', 100, '2024-07-01 10:01'),
  (3,  'pay_A', 'succeeded', 100, '2024-07-01 10:01'),
  (4,  'pay_B', 'succeeded', 250, '2024-07-01 11:00'),
  (5,  'pay_C', 'succeeded', 80,  '2024-07-01 12:00'),
  (6,  'pay_C', 'refunded',  80,  '2024-07-03 09:00'),
  (7,  'pay_D', 'pending',   40,  '2024-07-02 08:00'),
  (8,  'pay_D', 'failed',    40,  '2024-07-02 08:05'),
  (9,  'pay_E', 'succeeded', 300, '2024-07-02 14:00'),
  (10, 'pay_F', 'succeeded', 55,  '2024-07-02 15:00'),
  (11, 'pay_G', 'pending',   70,  '2024-07-02 16:00'),
  (12, 'pay_G', 'succeeded', 70,  '2024-07-02 16:00');

INSERT INTO ledger VALUES
  (1, 'pay_A', 100, '2024-07-01 10:02'),
  (2, 'pay_B', 205, '2024-07-01 11:01'),
  (3, 'pay_C', 80,  '2024-07-01 12:01'),
  (4, 'pay_D', 40,  '2024-07-02 08:01'),
  (5, 'pay_E', 150, '2024-07-02 14:01'),
  (6, 'pay_E', 150, '2024-07-02 14:02'),
  (7, NULL,    500, '2024-07-02 18:00'),
  (8, NULL,    20,  '2024-07-03 10:00'),
  (9, 'pay_H', 90,  '2024-07-03 11:00'),
  (10, 'pay_G', 70, '2024-07-02 16:01');
