package namespaces

import (
	"context"

	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
)

// NamespaceService is the interface that provides the namespace service methods.
type NamespaceService interface {
	// Create creates a new namespace.
	Create(context.Context, *models.Namespace) (*models.Namespace, error)
	// Get gets the namespace with the specified namespace ID.
	Get(context.Context, uint) (*models.Namespace, error)
	// GetByName gets the namespace with the specified namespace name.
	GetByName(context.Context, string) (*models.Namespace, error)
	// ListNamespace lists all namespaces.
	ListNamespace(ctx context.Context) ([]*models.Namespace, error)
}

type namespaceService struct {
	tx *query.Query
}

// NewNamespaceService creates a new namespace service.
func NewNamespaceService(txs ...*query.Query) NamespaceService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &namespaceService{
		tx: tx,
	}
}

// Create creates a new namespace.
func (s *namespaceService) Create(ctx context.Context, namespace *models.Namespace) (*models.Namespace, error) {
	err := s.tx.Namespace.WithContext(ctx).Create(namespace)
	if err != nil {
		return nil, err
	}
	return namespace, nil
}

// Get gets the namespace with the specified namespace ID.
func (s *namespaceService) Get(ctx context.Context, id uint) (*models.Namespace, error) {
	ns, err := s.tx.Namespace.WithContext(ctx).Where(s.tx.Namespace.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return ns, nil
}

// GetByName gets the namespace with the specified namespace name.
func (s *namespaceService) GetByName(ctx context.Context, name string) (*models.Namespace, error) {
	ns, err := s.tx.Namespace.WithContext(ctx).Where(s.tx.Namespace.Name.Eq(name)).First()
	if err != nil {
		return nil, err
	}
	return ns, nil
}

// ListNamespace lists all namespaces.
func (s *namespaceService) ListNamespace(ctx context.Context) ([]*models.Namespace, error) {
	return s.tx.Namespace.WithContext(ctx).Find()
}
