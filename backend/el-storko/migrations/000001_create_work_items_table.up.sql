CREATE TABLE IF NOT EXISTS work_items (
    id SERIAL PRIMARY KEY,
    type TEXT NOT NULL CHECK (type IN ('epic', 'task')),
    parent_id INTEGER REFERENCES work_items(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'todo' CHECK (status IN ('todo', 'in_progress', 'blocked', 'done')),
    source TEXT NOT NULL DEFAULT 'personal' CHECK (source IN ('personal', 'jira', 'agent')),
    jira_key TEXT UNIQUE,
    jira_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_work_items_parent_id ON work_items(parent_id);
CREATE INDEX idx_work_items_source ON work_items(source);
CREATE INDEX idx_work_items_status ON work_items(status);
