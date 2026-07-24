package work_item

import "github.com/shadow-dom/solid-fiber/pkg/storage"

// Migrations is the ordered, forward-only schema history for the work_item
// domain. Add a new entry (never edit an applied one) to evolve the schema.
var Migrations = []storage.Migration{
	{
		Version: 1,
		Name:    "create_work_items",
		SQL: `
CREATE TABLE IF NOT EXISTS work_items (
	id                   TEXT PRIMARY KEY,
	title                TEXT    NOT NULL,
	description_markdown TEXT    NOT NULL DEFAULT '',
	parent_id            TEXT    NOT NULL DEFAULT '',
	column_id            TEXT    NOT NULL DEFAULT '',
	assignee_id          TEXT    NOT NULL DEFAULT '',
	reporter_id          TEXT    NOT NULL DEFAULT '',
	sprint_id            TEXT    NOT NULL DEFAULT '',
	priority             INTEGER NOT NULL DEFAULT 0,
	estimate_hours       REAL    NOT NULL DEFAULT 0,
	story_points         REAL    NOT NULL DEFAULT 0,
	due_date             INTEGER NOT NULL DEFAULT 0,
	is_milestone         INTEGER NOT NULL DEFAULT 0,
	epic_color           TEXT    NOT NULL DEFAULT '',
	labels               TEXT    NOT NULL DEFAULT '[]',
	project_id           TEXT    NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_work_items_project_id ON work_items(project_id);
`,
	},
}
