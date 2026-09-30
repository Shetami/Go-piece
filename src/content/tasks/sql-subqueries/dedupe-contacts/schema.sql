CREATE TABLE contacts (
  id         int PRIMARY KEY,
  email      text,               -- NULL: контакт без почты (только телефон)
  name       text NOT NULL,
  updated_at timestamp           -- NULL: импорт из старой CRM, дата неизвестна
);

INSERT INTO contacts VALUES
  (1,  'anna@mail.ru',     'Анна',       '2024-01-10 10:00'),
  (2,  'Anna@Mail.ru ',    'Анна К.',    '2024-03-01 09:00'),
  (3,  ' ANNA@mail.ru',    'Анна Котова', NULL),
  (4,  'boris@ya.ru',      'Борис',      '2024-02-02 12:00'),
  (5,  'boris@ya.ru',      'Борис П.',   '2024-02-02 12:00'),
  (6,  'vera@gmail.com',   'Вера',        NULL),
  (7,  'Vera@gmail.com',   'Вера С.',     NULL),
  (8,  NULL,               'Глеб',       '2024-01-01 00:00'),
  (9,  NULL,               'Глеб Р.',    '2024-01-02 00:00'),
  (10, 'dina@list.ru',     'Дина',       '2023-12-12 12:00');
