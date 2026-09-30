CREATE TABLE subscribers (
  id              int PRIMARY KEY,
  email           text NOT NULL,
  unsubscribed_at timestamp          -- NULL: подписка активна
);

-- subscriber_id пуст, если подписчика удалили по запросу (GDPR),
-- а строку об отправке сохранили для статистики
CREATE TABLE sends (
  id            int PRIMARY KEY,
  campaign_id   int NOT NULL,
  subscriber_id int REFERENCES subscribers (id),
  sent_at       timestamp NOT NULL
);

INSERT INTO subscribers VALUES
  (1, 'anna@example.com',  NULL),
  (2, 'boris@example.com', NULL),
  (3, 'vera@example.com',  NULL),
  (4, 'gleb@example.com',  '2024-03-01 12:00'),
  (5, 'dina@example.com',  NULL),
  (6, 'egor@example.com',  NULL);

INSERT INTO sends VALUES
  (1, 7, 1,    '2024-03-10 09:00'),
  (2, 7, NULL, '2024-03-10 09:00'),
  (3, 7, 3,    '2024-03-10 09:01'),
  (4, 5, 2,    '2024-02-01 09:00'),
  (5, 5, 5,    '2024-02-01 09:00'),
  (6, 7, 3,    '2024-03-11 09:00');
