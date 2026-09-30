CREATE TABLE users (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE campaigns (
  id      int PRIMARY KEY,
  name    text NOT NULL,
  channel text NOT NULL
);

-- Рекламные касания: клики и показы
CREATE TABLE touches (
  user_id     int NOT NULL REFERENCES users(id),
  campaign_id int NOT NULL REFERENCES campaigns(id),
  kind        text NOT NULL, -- click, impression
  ts          timestamp NOT NULL
);

CREATE TABLE orders (
  id         int PRIMARY KEY,
  user_id    int NOT NULL REFERENCES users(id),
  amount     numeric(10,2) NOT NULL,
  status     text NOT NULL, -- paid, cancelled
  created_at timestamp NOT NULL
);

INSERT INTO users VALUES
  (1, 'Аня'), (2, 'Боря'), (3, 'Вера'), (4, 'Гоша'), (5, 'Дина');

INSERT INTO campaigns VALUES
  (1, 'Директ: бренд',    'search'),
  (2, 'VK: ретаргетинг',  'social'),
  (3, 'Рассылка: июнь',   'email'),
  (4, 'Кешбэк-сервис',    'affiliate'),
  (5, 'Директ: товары',   'search');

INSERT INTO touches VALUES
  (1, 3, 'click',      '2024-06-01 10:00'),
  (1, 1, 'click',      '2024-06-03 12:00'),
  (1, 2, 'click',      '2024-06-06 08:00'),
  (2, 2, 'click',      '2024-06-01 09:00'),
  (2, 5, 'impression', '2024-06-02 09:00'),
  (3, 3, 'click',      '2024-06-01 09:00'),
  (4, 2, 'click',      '2024-06-10 12:00'),
  (4, 5, 'click',      '2024-06-10 12:00'),
  (4, 4, 'impression', '2024-06-10 12:20'),
  (5, 4, 'click',      '2024-06-15 10:00');

INSERT INTO orders VALUES
  (1, 1, 3000.00, 'paid',      '2024-06-05 20:00'),
  (2, 1, 1000.00, 'paid',      '2024-06-20 11:00'),
  (3, 2, 2500.00, 'paid',      '2024-06-08 09:00'),
  (4, 3, 4000.00, 'paid',      '2024-06-08 09:00:01'),
  (5, 4, 1500.00, 'paid',      '2024-06-10 12:30'),
  (6, 4, 9000.00, 'cancelled', '2024-06-11 10:00'),
  (7, 5,  500.00, 'paid',      '2024-06-14 18:00'),
  (8, 4,  700.00, 'paid',      '2024-06-12 09:00');
