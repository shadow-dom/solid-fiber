package work_item

type WorkItem struct {
	ID                  string   `json:"id"`
	Title               string   `json:"title"`
	DescriptionMarkDown string   `json:"description_markdown"`
	ParentID            int      `json:"parent_id,omitempty"`
	ColumnID            int      `json:"column_id,omitempty"`
	AssigneeID          int      `json:"assignee_id,omitempty"`
	ReporterID          int      `json:"reporter_id,omitempty"`
	SprintID            int      `json:"sprint_id,omitempty"`
	Priority            int      `json:"priority,omitempty"`
	EstimateHours       float64  `json:"estimate_hours,omitempty"`
	StoryPoints         float64  `json:"story_points,omitempty"`
	DueDate             int64    `json:"due_date,omitempty"`
	IsMilestone         bool     `json:"is_milestone,omitempty"`
	EpicColor           string   `json:"epic_color,omitempty"`
	Labels              []string `json:"labels,omitempty"`
	ProjectID           int      `json:"project_id"`
}
