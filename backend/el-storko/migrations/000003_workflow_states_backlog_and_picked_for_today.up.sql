ALTER TABLE work_items DROP CONSTRAINT work_items_status_check;

UPDATE work_items SET status = 'backlog' WHERE status = 'todo';

ALTER TABLE work_items
  ADD CONSTRAINT work_items_status_check
  CHECK (status IN ('backlog', 'picked_for_today', 'in_progress', 'blocked', 'done'));

ALTER TABLE work_items ALTER COLUMN status SET DEFAULT 'backlog';
