-- Стоимость доставки каждой посылки.
-- Колонки: id, billable_kg (round(…, 2)), cost.
-- Порядок: по id.
SELECT p.id,
       greatest(p.weight_kg, p.length_cm * p.width_cm * p.height_cm / 5000) AS billable_kg,
       t.base + t.per_kg * (greatest(p.weight_kg, p.length_cm * p.width_cm * p.height_cm / 5000) - 1) AS cost
FROM parcels p
JOIN tariffs t ON t.zone = p.zone
ORDER BY p.id;
