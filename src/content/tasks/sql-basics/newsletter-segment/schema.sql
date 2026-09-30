CREATE TABLE customers (
  id         int PRIMARY KEY,
  email      text,
  country    text,        -- ISO-код; NULL — страна не указана
  is_blocked boolean      -- NULL — флаг не заполняли, клиент не заблокирован
);

-- Жёсткие отказы почтового сервера. Парсер логов иногда не вытаскивает адрес.
CREATE TABLE bounces (
  email       text,
  bounced_at  timestamp NOT NULL
);

INSERT INTO customers VALUES
  (1,  'anna@mail.ru',        'RU', false),
  (2,  ' Boris@Gmail.com ',   'RU', NULL),
  (3,  'vera@yandex.ru',      NULL, NULL),
  (4,  '',                    'RU', false),
  (5,  NULL,                  'RU', false),
  (6,  'gosha@mail.ru',       'BY', false),
  (7,  'dina@mail.ru',        'KZ', true),
  (8,  'boris@gmail.com',     'RU', false),
  (9,  'EGOR@corp.com',       'RU', false),
  (10, 'zhanna@corp.com',     'AM', NULL),
  (11, '   ',                 'RU', false),
  (12, 'ivan@mail.ru',        'RU', false);

INSERT INTO bounces VALUES
  ('egor@corp.com ',  '2024-04-02 10:00'),
  (NULL,              '2024-04-03 11:30'),
  ('Oleg@mail.ru',    '2024-04-05 09:15');
