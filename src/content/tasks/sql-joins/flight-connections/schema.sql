CREATE TABLE flights (
  id        int PRIMARY KEY,
  flight_no text      NOT NULL,
  origin    text      NOT NULL,
  dest      text      NOT NULL,
  dep_at    timestamp NOT NULL,
  arr_at    timestamp NOT NULL,
  status    text      NOT NULL  -- 'scheduled', 'cancelled'
);

-- Пассажир прилетает рейсом inbound_flight_id и хочет дальше, в final_dest
CREATE TABLE transfers (
  id                int PRIMARY KEY,
  passenger         text NOT NULL,
  inbound_flight_id int  NOT NULL REFERENCES flights (id),
  final_dest        text NOT NULL
);

INSERT INTO flights VALUES
  (1,  'SU100', 'LED', 'SVO', '2024-09-01 08:00', '2024-09-01 09:30', 'scheduled'),
  (2,  'SU200', 'KZN', 'SVO', '2024-09-01 20:00', '2024-09-01 21:40', 'scheduled'),
  (3,  'SU300', 'AER', 'VKO', '2024-09-01 07:00', '2024-09-01 09:30', 'scheduled'),
  (10, 'SU501', 'SVO', 'OVB', '2024-09-01 10:10', '2024-09-01 16:00', 'scheduled'),
  (11, 'SU503', 'SVO', 'OVB', '2024-09-01 10:20', '2024-09-01 16:10', 'scheduled'),
  (12, 'SU502', 'SVO', 'OVB', '2024-09-01 10:20', '2024-09-01 16:15', 'scheduled'),
  (13, 'SU505', 'SVO', 'OVB', '2024-09-01 09:50', '2024-09-01 15:40', 'scheduled'),
  (14, 'SU601', 'SVO', 'KJA', '2024-09-01 10:15', '2024-09-01 18:00', 'cancelled'),
  (15, 'SU603', 'SVO', 'KJA', '2024-09-01 21:30', '2024-09-02 05:10', 'scheduled'),
  (16, 'SU701', 'VKO', 'OVB', '2024-09-01 10:00', '2024-09-01 16:00', 'scheduled'),
  (17, 'SU801', 'SVO', 'AER', '2024-09-02 09:40', '2024-09-02 12:00', 'scheduled'),
  (18, 'SU803', 'SVO', 'AER', '2024-09-02 09:41', '2024-09-02 12:01', 'scheduled'),
  (19, 'SU901', 'SVO', 'MRV', '2024-09-01 10:00', '2024-09-01 12:30', 'scheduled');

INSERT INTO transfers VALUES
  (1, 'Анна',  1, 'OVB'),
  (2, 'Борис', 1, 'KJA'),
  (3, 'Вера',  2, 'AER'),
  (4, 'Глеб',  3, 'OVB'),
  (5, 'Дина',  1, 'MRV');
