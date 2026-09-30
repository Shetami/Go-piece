SELECT p.id,
       round(b.kg, 2)                                   AS billable_kg,
       t.base + t.per_kg * ceil(greatest(b.kg - 1, 0))  AS cost
FROM parcels p
CROSS JOIN LATERAL (
  -- 5000.0, а не 5000: иначе деление целых отбросит дробную часть.
  -- greatest в Postgres пропускает NULL, так что хватает любой из двух оценок.
  SELECT greatest(p.weight_kg, p.length_cm * p.width_cm * p.height_cm / 5000.0) AS kg
) b
JOIN tariffs t ON t.zone = coalesce(p.zone, 3)
WHERE b.kg IS NOT NULL
ORDER BY p.id;
