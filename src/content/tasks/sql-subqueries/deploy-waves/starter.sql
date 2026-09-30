-- Волна выкатки: без зависимостей — 1, иначе 1 + максимальная волна среди зависимостей.
-- Сервис в цикле или зависящий (через любое число шагов) от цикла — wave пустая.
-- Колонки: service, wave. Порядок: wave (пустые — в конце), затем service.
SELECT s.name AS service,
       1 + (SELECT count(*) FROM deps d WHERE d.service = s.name) AS wave
FROM services s
ORDER BY wave NULLS LAST, service;
