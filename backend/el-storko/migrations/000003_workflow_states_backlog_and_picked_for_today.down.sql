ALTER TABLE work_items DROP CONSTRAINT work_items_status_check;

UPDATE work_items SET status = 'todo' WHERE status IN ('backlog', 'picked_for_today');

ALTER TABLE work_items
  ADD CONSTRAINT work_items_status_check
  CHECK (status IN ('todo', 'in_progress', 'blocked', 'done'));

ALTER TABLE work_items ALTER COLUMN status SET DEFAULT 'todo';
