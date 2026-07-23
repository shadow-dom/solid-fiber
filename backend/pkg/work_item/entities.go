package work_item

// WorkItem is the core domain entity. Identity and all foreign-key references
// use string UUIDs for consistency across the system.
type WorkItem struct {
	ID                  string   `json:"id"`
	Title               string   `json:"title"`
	DescriptionMarkDown string   `json:"description_markdown"`
	ParentID            string   `json:"parent_id,omitempty"`
	ColumnID            string   `json:"column_id,omitempty"`
	AssigneeID          string   `json:"assignee_id,omitempty"`
	ReporterID          string   `json:"reporter_id,omitempty"`
	SprintID            string   `json:"sprint_id,omitempty"`
	Priority            int      `json:"priority,omitempty"`
	EstimateHours       float64  `json:"estimate_hours,omitempty"`
	StoryPoints         float64  `json:"story_points,omitempty"`
	DueDate             int64    `json:"due_date,omitempty"`
	IsMilestone         bool     `json:"is_milestone,omitempty"`
	EpicColor           string   `json:"epic_color,omitempty"`
	Labels              []string `json:"labels,omitempty"`
	ProjectID           string   `json:"project_id"`
}
