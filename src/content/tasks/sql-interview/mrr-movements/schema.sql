CREATE TABLE plans (
  id            int PRIMARY KEY,
  name          text NOT NULL,
  monthly_price int NOT NULL
);

CREATE TABLE customers (
  id   int PRIMARY KEY,
  name text NOT NULL
);

-- Период подписки: [started_on, ended_on). ended_on IS NULL — действует.
-- Смена тарифа — это конец одной строки и начало следующей.
CREATE TABLE subscriptions (
  id          int PRIMARY KEY,
  customer_id int NOT NULL REFERENCES customers(id),
  plan_id     int NOT NULL REFERENCES plans(id),
  started_on  date NOT NULL,
  ended_on    date
);

INSERT INTO plans VALUES
  (1, 'Basic', 10),
  (2, 'Pro',   30),
  (3, 'Team',  50);

INSERT INTO customers VALUES
  (1, 'Альфа'),
  (2, 'Бета'),
  (3, 'Гамма'),
  (4, 'Дельта'),
  (5, 'Эпсилон');

INSERT INTO subscriptions VALUES
  (1, 1, 1, '2023-12-15', '2024-03-10'),
  (2, 1, 2, '2024-03-10', NULL),
  (3, 2, 3, '2024-01-01', '2024-04-01'),
  (4, 2, 1, '2024-04-01', NULL),
  (5, 3, 2, '2024-01-20', '2024-03-15'),
  (6, 3, 1, '2024-05-01', NULL),
  (7, 4, 1, '2024-02-01', '2024-02-20'),
  (8, 5, 3, '2024-03-02', NULL);
