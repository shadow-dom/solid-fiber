package work_item

import (
	"errors"

	"github.com/google/uuid"
)

// ErrTitleRequired is returned when a work item is created or updated without a title.
var ErrTitleRequired = errors.New("title is required")

// Service is the application-facing port for work item use cases.
type Service interface {
	CreateWorkItem(workItem *WorkItem) (*WorkItem, error)
	GetWorkItemByID(id string) (*WorkItem, error)
	UpdateWorkItem(workItem *WorkItem) (*WorkItem, error)
	DeleteWorkItem(id string) error
	// ListWorkItemsByProjectID returns a page of items (ordered by id, sliced by
	// offset/limit) along with the total count for the project.
	ListWorkItemsByProjectID(projectID string, limit, offset int) ([]*WorkItem, int, error)
}

type service struct {
	repo Repository
}

// NewService returns a Service backed by the given Repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateWorkItem(workItem *WorkItem) (*WorkItem, error) {
	if workItem.Title == "" {
		return nil, ErrTitleRequired
	}
	// IDs are assigned server-side; any client-supplied ID is ignored.
	workItem.ID = uuid.NewString()
	return s.repo.Create(workItem)
}

func (s *service) GetWorkItemByID(id string) (*WorkItem, error) {
	return s.repo.GetByID(id)
}

func (s *service) UpdateWorkItem(workItem *WorkItem) (*WorkItem, error) {
	if workItem.Title == "" {
		return nil, ErrTitleRequired
	}
	return s.repo.Update(workItem)
}

func (s *service) DeleteWorkItem(id string) error {
	return s.repo.Delete(id)
}

func (s *service) ListWorkItemsByProjectID(projectID string, limit, offset int) ([]*WorkItem, int, error) {
	items, err := s.repo.ListByProjectID(projectID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountByProjectID(projectID)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
