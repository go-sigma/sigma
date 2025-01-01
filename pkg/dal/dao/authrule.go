package dao

import (
	"context"

	"gorm.io/gen/field"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

//go:generate mockgen -destination=mocks/authrule.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao AuthRuleService
//go:generate mockgen -destination=mocks/authrule_factory.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao AuthRuleServiceFactory

// AuthRuleService is the interface that provides methods to operate on auth rule model
type AuthRuleService interface {
	// Create creates a new auth rule record in the database
	Create(ctx context.Context, authRule *models.AuthRule) error
	// ListByResource list auth rule by scope
	ListByScope(ctx context.Context, scopes []ScopeItem) ([]*models.AuthRule, error)
}

type authRuleService struct {
	tx *query.Query
}

// AuthRuleServiceFactory is the interface that provides the auth rule service factory methods
type AuthRuleServiceFactory interface {
	New(txs ...*query.Query) AuthRuleService
}

type authRuleServiceFactory struct{}

// NewAuthRuleServiceFactory creates a new auth rule service factory
func NewAuthRuleServiceFactory() AuthRuleServiceFactory {
	return &authRuleServiceFactory{}
}

// New creates a new auth rule service
func (s *authRuleServiceFactory) New(txs ...*query.Query) AuthRuleService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &authRuleService{
		tx: tx,
	}
}

// Create create a auth rule
func (s *authRuleService) Create(ctx context.Context, authRule *models.AuthRule) error {
	return s.tx.AuthRule.WithContext(ctx).Create(authRule)
}

// ScopeItem scope item
type ScopeItem struct {
	ScopeType  enums.AuthScope `json:"scope_type"`
	ScopeValue string          `json:"scope_value"`
}

// ListByResource list auth rule by scope
func (s *authRuleService) ListByScope(ctx context.Context, scopes []ScopeItem) ([]*models.AuthRule, error) {
	if len(scopes) == 0 {
		return nil, nil
	}
	var conds = make([]field.Expr, 0, len(scopes))
	for _, scope := range scopes {
		conds = append(conds, field.And(
			s.tx.AuthRule.ScopeType.Eq(scope.ScopeType),
			s.tx.AuthRule.ScopeValue.Eq(scope.ScopeValue),
		))
	}
	return s.tx.AuthRule.WithContext(ctx).Where(field.Or(conds...)).Preload(s.tx.AuthRule.Role).Find()
}
