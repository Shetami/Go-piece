SELECT player,
       row_number() OVER (ORDER BY points DESC, player) AS rn,
       rank()       OVER (ORDER BY points DESC)         AS rnk,
       dense_rank() OVER (ORDER BY points DESC)         AS drnk
FROM scores
ORDER BY points DESC, player;
