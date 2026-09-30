WITH ordered AS (
  SELECT customer, starts_on, ends_on,
         max(ends_on) OVER (PARTITION BY customer ORDER BY starts_on, ends_on
                            ROWS BETWEEN UNBOUNDED PRECEDING AND 1 PRECEDING) AS prev_max_end
  FROM periods
),
flagged AS (
  SELECT *,
         CASE WHEN prev_max_end IS NULL OR starts_on > prev_max_end + 1 THEN 1 ELSE 0 END AS is_start
  FROM ordered
),
grouped AS (
  SELECT *,
         sum(is_start) OVER (PARTITION BY customer ORDER BY starts_on, ends_on
                             ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS grp
  FROM flagged
)
SELECT customer,
       min(starts_on)                    AS starts_on,
       max(ends_on)                      AS ends_on,
       max(ends_on) - min(starts_on) + 1 AS days
FROM grouped
GROUP BY customer, grp
ORDER BY customer, starts_on;
