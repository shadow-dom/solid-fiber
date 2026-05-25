package work_item

type Service interface {
	CreateWorkItem(workItem *WorkItem) (*WorkItem, error)
	GetWorkItemByID(id int) (*WorkItem, error)
	UpdateWorkItem(workItem *WorkItem) (*WorkItem, error)
	DeleteWorkItem(id int) error
	ListWorkItemsByProjectID(projectID int) ([]*WorkItem, error)
}
