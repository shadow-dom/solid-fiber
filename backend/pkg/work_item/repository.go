package work_item

import (
	"errors"
	"sync"
)

// ErrNotFound is returned when a work item does not exist.
var ErrNotFound = errors.New("work item not found")

// Repository is the persistence port for work items.
type Repository interface {
	Create(item *WorkItem) (*WorkItem, error)
	GetByID(id string) (*WorkItem, error)
	Update(item *WorkItem) (*WorkItem, error)
	Delete(id string) error
	ListByProjectID(projectID string) ([]*WorkItem, error)
}

// inMemoryRepository is a goroutine-safe, in-memory Repository implementation.
// It is intended as a starting point until a real datastore is wired in.
type inMemoryRepository struct {
	mu    sync.RWMutex
	items map[string]*WorkItem
}

// NewInMemoryRepository returns an empty in-memory Repository.
func NewInMemoryRepository() Repository {
	return &inMemoryRepository{items: make(map[string]*WorkItem)}
}

func (r *inMemoryRepository) Create(item *WorkItem) (*WorkItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := *item
	r.items[stored.ID] = &stored
	return &stored, nil
}

func (r *inMemoryRepository) GetByID(id string) (*WorkItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	clone := *item
	return &clone, nil
}

func (r *inMemoryRepository) Update(item *WorkItem) (*WorkItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[item.ID]; !ok {
		return nil, ErrNotFound
	}
	stored := *item
	r.items[stored.ID] = &stored
	return &stored, nil
}

func (r *inMemoryRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return ErrNotFound
	}
	delete(r.items, id)
	return nil
}

func (r *inMemoryRepository) ListByProjectID(projectID string) ([]*WorkItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]*WorkItem, 0)
	for _, item := range r.items {
		if item.ProjectID == projectID {
			clone := *item
			items = append(items, &clone)
		}
	}
	return items, nil
}
