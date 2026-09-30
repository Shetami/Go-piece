CREATE TABLE customers (
  id   int PRIMARY KEY,
  name text NOT NULL
);

-- ended_on — последний день, когда подписка ещё действует (включительно); NULL — бессрочно
CREATE TABLE subscriptions (
  id          int PRIMARY KEY,
  customer_id int  NOT NULL REFERENCES customers (id),
  plan        text NOT NULL,
  started_on  date NOT NULL,
  ended_on    date
);

-- Счёт за месяц: period_start — первое число оплачиваемого месяца.
-- paid_at пуст, пока счёт не оплачен. subscription_id пуст у разовых счетов
-- за услуги вне подписки.
CREATE TABLE invoices (
  id              int PRIMARY KEY,
  subscription_id int REFERENCES subscriptions (id),
  period_start    date NOT NULL,
  amount          numeric NOT NULL,
  paid_at         timestamp
);

INSERT INTO customers VALUES
  (1, 'Альфа'),
  (2, 'Бета'),
  (3, 'Гамма'),
  (4, 'Дельта'),
  (5, 'Омега');

INSERT INTO subscriptions VALUES
  (10, 1, 'pro',   '2024-01-01', NULL),
  (11, 1, 'extra', '2024-05-15', NULL),
  (12, 2, 'basic', '2024-02-01', '2024-05-31'),
  (13, 3, 'pro',   '2024-06-30', NULL),
  (14, 3, 'basic', '2024-07-01', NULL),
  (15, 4, 'pro',   '2023-11-01', '2024-06-10'),
  (16, 5, 'basic', '2024-03-01', NULL),
  (17, 2, 'pro',   '2024-06-01', NULL);

INSERT INTO invoices VALUES
  (1,  10,   '2024-06-01', 100, '2024-06-02 10:00'),
  (2,  11,   '2024-05-01', 50,  '2024-05-16 09:00'),
  (3,  11,   '2024-06-01', 50,  NULL),
  (4,  15,   '2024-05-01', 100, '2024-05-03 12:00'),
  (5,  16,   '2024-06-01', 30,  NULL),
  (6,  16,   '2024-06-01', 30,  '2024-06-20 18:00'),
  (7,  NULL, '2024-06-01', 500, NULL),
  (8,  17,   '2024-07-01', 100, '2024-06-28 11:00');
