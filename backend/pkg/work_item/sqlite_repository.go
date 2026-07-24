package work_item

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/shadow-dom/solid-fiber/pkg/storage"
)

// The column list shared by reads and writes, in a stable order.
const workItemColumns = `id, title, description_markdown, parent_id, column_id,
	assignee_id, reporter_id, sprint_id, priority, estimate_hours, story_points,
	due_date, is_milestone, epic_color, labels, project_id`

// sqliteRepository is a Repository backed by a SQLite database.
type sqliteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository returns a Repository backed by db, applying any pending
// schema migrations first. Migration is idempotent, so this is safe to call on
// an already-migrated database.
func NewSQLiteRepository(db *sql.DB) (Repository, error) {
	if err := storage.Migrate(db, Migrations); err != nil {
		return nil, fmt.Errorf("migrate work_items schema: %w", err)
	}
	return &sqliteRepository{db: db}, nil
}

func (r *sqliteRepository) Create(item *WorkItem) (*WorkItem, error) {
	labels, err := marshalLabels(item.Labels)
	if err != nil {
		return nil, err
	}
	_, err = r.db.Exec(
		`INSERT INTO work_items (`+workItemColumns+`)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.Title, item.DescriptionMarkDown, item.ParentID, item.ColumnID,
		item.AssigneeID, item.ReporterID, item.SprintID, item.Priority, item.EstimateHours,
		item.StoryPoints, item.DueDate, item.IsMilestone, item.EpicColor, labels, item.ProjectID,
	)
	if err != nil {
		return nil, fmt.Errorf("insert work item: %w", err)
	}
	clone := *item
	return &clone, nil
}

func (r *sqliteRepository) GetByID(id string) (*WorkItem, error) {
	row := r.db.QueryRow(
		`SELECT `+workItemColumns+` FROM work_items WHERE id = ?`, id)
	item, err := scanWorkItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get work item: %w", err)
	}
	return item, nil
}

func (r *sqliteRepository) Update(item *WorkItem) (*WorkItem, error) {
	labels, err := marshalLabels(item.Labels)
	if err != nil {
		return nil, err
	}
	res, err := r.db.Exec(
		`UPDATE work_items SET
			title = ?, description_markdown = ?, parent_id = ?, column_id = ?,
			assignee_id = ?, reporter_id = ?, sprint_id = ?, priority = ?,
			estimate_hours = ?, story_points = ?, due_date = ?, is_milestone = ?,
			epic_color = ?, labels = ?, project_id = ?
		 WHERE id = ?`,
		item.Title, item.DescriptionMarkDown, item.ParentID, item.ColumnID,
		item.AssigneeID, item.ReporterID, item.SprintID, item.Priority,
		item.EstimateHours, item.StoryPoints, item.DueDate, item.IsMilestone,
		item.EpicColor, labels, item.ProjectID, item.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("update work item: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrNotFound
	}
	clone := *item
	return &clone, nil
}

func (r *sqliteRepository) Delete(id string) error {
	res, err := r.db.Exec(`DELETE FROM work_items WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete work item: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sqliteRepository) ListByProjectID(projectID string, limit, offset int) ([]*WorkItem, error) {
	if limit <= 0 {
		limit = -1 // SQLite: no limit
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := r.db.Query(
		`SELECT `+workItemColumns+` FROM work_items WHERE project_id = ? ORDER BY id LIMIT ? OFFSET ?`,
		projectID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list work items: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items := make([]*WorkItem, 0)
	for rows.Next() {
		item, err := scanWorkItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan work item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate work items: %w", err)
	}
	return items, nil
}

func (r *sqliteRepository) CountByProjectID(projectID string) (int, error) {
	var n int
	if err := r.db.QueryRow(
		`SELECT COUNT(*) FROM work_items WHERE project_id = ?`, projectID,
	).Scan(&n); err != nil {
		return 0, fmt.Errorf("count work items: %w", err)
	}
	return n, nil
}

// scanner is satisfied by both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanWorkItem(s scanner) (*WorkItem, error) {
	var item WorkItem
	var labels string
	if err := s.Scan(
		&item.ID, &item.Title, &item.DescriptionMarkDown, &item.ParentID, &item.ColumnID,
		&item.AssigneeID, &item.ReporterID, &item.SprintID, &item.Priority, &item.EstimateHours,
		&item.StoryPoints, &item.DueDate, &item.IsMilestone, &item.EpicColor, &labels, &item.ProjectID,
	); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(labels), &item.Labels); err != nil {
		return nil, fmt.Errorf("decode labels: %w", err)
	}
	return &item, nil
}

func marshalLabels(labels []string) (string, error) {
	if labels == nil {
		return "[]", nil
	}
	b, err := json.Marshal(labels)
	if err != nil {
		return "", fmt.Errorf("encode labels: %w", err)
	}
	return string(b), nil
}
