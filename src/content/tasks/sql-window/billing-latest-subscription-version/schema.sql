CREATE TABLE plans (
  code          text PRIMARY KEY,
  name          text          NOT NULL,
  monthly_price numeric(10,2) NOT NULL
);

-- Каждое изменение подписки пишет новую версию: история, а не текущее состояние.
CREATE TABLE subscription_versions (
  id          int PRIMARY KEY,
  customer_id int       NOT NULL,
  plan_code   text      NOT NULL REFERENCES plans (code),
  status      text      NOT NULL CHECK (status IN ('active', 'cancelled')),
  updated_at  timestamp NOT NULL
);

INSERT INTO plans VALUES
  ('basic',  'Basic',  10.00),
  ('pro',    'Pro',    30.00),
  ('team',   'Team',  100.00),
  ('legacy', 'Legacy',  5.00);

INSERT INTO subscription_versions VALUES
  -- 1: basic → pro
  ( 1, 1, 'basic', 'active',    '2024-01-01 10:00'),
  ( 2, 1, 'pro',   'active',    '2024-02-01 10:00'),
  -- 2: ушёл
  ( 3, 2, 'pro',   'active',    '2024-01-10 10:00'),
  ( 4, 2, 'pro',   'cancelled', '2024-03-01 10:00'),
  -- 3: две версии в одну секунду — действует та, у которой id больше
  ( 5, 3, 'basic', 'active',    '2024-03-05 10:00'),
  ( 6, 3, 'team',  'active',    '2024-03-05 10:00'),
  -- 4: отменял, потом вернулся
  ( 7, 4, 'team',  'cancelled', '2024-01-01 10:00'),
  ( 8, 4, 'basic', 'active',    '2024-02-15 10:00'),
  -- 5: одна версия
  ( 9, 5, 'pro',   'active',    '2024-02-20 10:00'),
  -- 6: версии записаны не по порядку id
  (11, 6, 'team',  'active',    '2024-03-10 10:00'),
  (10, 6, 'basic', 'active',    '2024-03-11 10:00');
