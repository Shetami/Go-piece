CREATE TABLE teams (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE sla (
  priority      text PRIMARY KEY,
  first_reply_h int NOT NULL -- рабочих часов на первый ответ
);

CREATE TABLE tickets (
  id         int PRIMARY KEY,
  team_id    int NOT NULL REFERENCES teams(id),
  priority   text NOT NULL REFERENCES sla(priority),
  created_at timestamp NOT NULL
);

-- Сообщения в тикете: от оператора, от клиента или автоответ бота
CREATE TABLE messages (
  ticket_id  int NOT NULL REFERENCES tickets(id),
  author     text NOT NULL, -- agent, customer, bot
  created_at timestamp NOT NULL
);

-- Нерабочие дни помимо выходных
CREATE TABLE holidays (
  day date PRIMARY KEY
);

INSERT INTO teams VALUES
  (1, 'Платежи'), (2, 'Доставка'), (3, 'Партнёры');

INSERT INTO sla VALUES
  ('high', 2), ('normal', 8);

INSERT INTO holidays VALUES ('2024-07-03');

-- 2024-07-01 — понедельник
INSERT INTO tickets VALUES
  (1,  1, 'high',   '2024-07-01 09:00'),
  (2,  1, 'high',   '2024-07-02 18:00'),
  (3,  1, 'normal', '2024-07-05 18:30'),
  (4,  1, 'high',   '2024-07-01 15:00'),
  (5,  1, 'high',   '2024-07-01 16:00'),
  (6,  2, 'normal', '2024-07-06 12:00'),
  (7,  2, 'normal', '2024-07-02 10:00'),
  (8,  2, 'high',   '2024-07-05 17:00'),
  (9,  2, 'normal', '2024-07-08 11:00'),
  (10, 2, 'normal', '2024-07-04 08:00'),
  (11, 1, 'normal', '2024-07-03 14:00');

INSERT INTO messages VALUES
  (1,  'bot',      '2024-07-01 09:00:05'),
  (1,  'agent',    '2024-07-01 12:10'),
  (1,  'agent',    '2024-07-01 11:30'),
  (2,  'agent',    '2024-07-04 10:30'),
  (3,  'customer', '2024-07-06 10:00'),
  (3,  'agent',    '2024-07-08 10:40'),
  (4,  'agent',    '2024-07-01 17:00'),
  (5,  'agent',    '2024-07-01 18:01'),
  (6,  'agent',    '2024-07-08 11:00'),
  (7,  'customer', '2024-07-02 10:05'),
  (7,  'agent',    '2024-07-04 10:00'),
  (8,  'bot',      '2024-07-05 17:00:03'),
  (10, 'agent',    '2024-07-04 09:50'),
  (11, 'agent',    '2024-07-04 10:15');
