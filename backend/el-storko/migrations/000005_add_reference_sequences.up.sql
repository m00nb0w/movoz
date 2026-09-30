CREATE SEQUENCE epic_reference_seq;
CREATE SEQUENCE task_reference_seq;

ALTER TABLE work_items ADD COLUMN reference_number INTEGER;

WITH numbered AS (
  SELECT id, row_number() OVER (PARTITION BY type ORDER BY id) AS rn
  FROM work_items
)
UPDATE work_items
SET reference_number = numbered.rn
FROM numbered
WHERE work_items.id = numbered.id;

SELECT setval('epic_reference_seq', COALESCE((SELECT MAX(reference_number) FROM work_items WHERE type = 'epic'), 0) + 1, false);
SELECT setval('task_reference_seq', COALESCE((SELECT MAX(reference_number) FROM work_items WHERE type = 'task'), 0) + 1, false);

ALTER TABLE work_items ALTER COLUMN reference_number SET NOT NULL;
