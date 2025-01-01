package dao

import (
	"context"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
)

//go:generate mockgen -destination=mocks/authrole.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao AuthRoleService
//go:generate mockgen -destination=mocks/authrole_factory.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao AuthRoleServiceFactory

// AuthRoleService is the interface that provides methods to operate on auth role model
type AuthRoleService interface {
	// Create creates a new auth role record in the database
	Create(ctx context.Context, authRole *models.AuthRole) error
}

type authRoleService struct {
	tx *query.Query
}

// AuthRoleServiceFactory is the interface that provides the auth role service factory methods
type AuthRoleServiceFactory interface {
	New(txs ...*query.Query) AuthRoleService
}

type authRoleServiceFactory struct{}

// NewAuthRoleServiceFactory creates a new auth role service factory
func NewAuthRoleServiceFactory() AuthRoleServiceFactory {
	return &authRoleServiceFactory{}
}

// New creates a new auth role service
func (s *authRoleServiceFactory) New(txs ...*query.Query) AuthRoleService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &authRoleService{
		tx: tx,
	}
}

// Create create a auth role
func (s *authRoleService) Create(ctx context.Context, authRole *models.AuthRole) error {
	return s.tx.AuthRole.WithContext(ctx).Create(authRole)
}
