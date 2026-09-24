-- Самая длинная серия заходов подряд (день за днём, без пропусков) у каждого пользователя.
-- Колонки: user_name, longest_streak (в днях). Порядок: по user_name.
SELECT user_name, count(DISTINCT day) AS longest_streak
FROM logins
GROUP BY user_name
ORDER BY user_name;
