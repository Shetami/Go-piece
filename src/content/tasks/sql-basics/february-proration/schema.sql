CREATE TABLE plans (
  id            int PRIMARY KEY,
  name          text NOT NULL,
  monthly_price int NOT NULL      -- ₽ за полный месяц
);

CREATE TABLE subscriptions (
  id         int PRIMARY KEY,
  customer   text NOT NULL,
  plan_id    int NOT NULL REFERENCES plans(id),
  started_on date NOT NULL,        -- первый оплачиваемый день
  ended_on   date                  -- последний оплачиваемый день, включительно; NULL — действует
);

INSERT INTO plans VALUES
  (1, 'basic', 990),
  (2, 'pro',   2900),
  (3, 'team',  4990);

INSERT INTO subscriptions VALUES
  (1, 'Альфа',  1, '2023-11-15', NULL),
  (2, 'Бета',   2, '2024-02-10', NULL),
  (3, 'Бета',   1, '2023-06-01', '2024-02-09'),
  (4, 'Гамма',  3, '2024-02-29', NULL),
  (5, 'Дельта', 2, '2024-01-15', '2024-01-31'),
  (6, 'Дельта', 1, '2024-03-01', NULL),
  (7, 'Гамма',  1, '2024-02-01', '2024-02-01'),
  (8, 'Омега',  2, '2024-01-20', '2024-03-10');
