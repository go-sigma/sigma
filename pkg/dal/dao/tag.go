package dao

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
	"github.com/ximager/ximager/pkg/types"
)

// TagService is the interface that provides the tag service methods.
type TagService interface {
	// Save save a new tag if conflict update.
	Save(ctx context.Context, tag *models.Tag) (*models.Tag, error)
	// Get gets the tag with the specified tag ID.
	GetByID(ctx context.Context, tagID uint) (*models.Tag, error)
	// GetByName gets the tag with the specified tag name.
	GetByName(context.Context, string, string) (*models.Tag, error)
	// DeleteByName deletes the tag with the specified tag name.
	DeleteByName(ctx context.Context, repository string, tag string) error
	// Incr increases the pull times of the artifact.
	Incr(ctx context.Context, id uint) error
	// ListByDtPagination lists the tags by the specified repository and pagination.
	ListByDtPagination(ctx context.Context, repository string, limit int, lastID ...uint) ([]*models.Tag, error)
	// ListTag lists the tags by the specified request.
	ListTag(ctx context.Context, req types.ListTagRequest) ([]*models.Tag, error)
	// CountArtifact counts the artifacts by the specified request.
	CountTag(ctx context.Context, req types.ListTagRequest) (int64, error)
	// DeleteByID deletes the tag with the specified tag ID.
	DeleteByID(ctx context.Context, id uint) error
}

type tagService struct {
	tx *query.Query
}

// NewTagService creates a new tag service.
func NewTagService(txs ...*query.Query) TagService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &tagService{
		tx: tx,
	}
}

// Save save a new tag if conflict update.
func (s *tagService) Save(ctx context.Context, tag *models.Tag) (*models.Tag, error) {
	err := s.tx.Tag.WithContext(ctx).Save(tag)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

// Get gets the tag with the specified tag ID.
func (s *tagService) GetByID(ctx context.Context, tagID uint) (*models.Tag, error) {
	tag, err := s.tx.Tag.WithContext(ctx).Where(s.tx.Tag.ID.Eq(tagID)).First()
	if err != nil {
		return nil, err
	}
	return tag, nil
}

// GetByName gets the tag with the specified tag name.
func (s *tagService) GetByName(ctx context.Context, repository, tag string) (*models.Tag, error) {
	tagObj, err := s.tx.Tag.WithContext(ctx).
		LeftJoin(s.tx.Repository, s.tx.Tag.RepositoryID.EqCol(s.tx.Repository.ID)).
		Where(s.tx.Tag.Name.Eq(tag)).
		Where(s.tx.Repository.Name.Eq(repository)).
		First()
	if err != nil {
		return nil, err
	}
	return tagObj, nil
}

// DeleteByName deletes the tag with the specified tag name.
func (s *tagService) DeleteByName(ctx context.Context, repository, tag string) error {
	matched, err := s.tx.Tag.WithContext(ctx).DeleteByName(repository, tag)
	if err != nil {
		return err
	}
	if matched == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Incr increases the pull times of the artifact.
func (s *tagService) Incr(ctx context.Context, id uint) error {
	_, err := s.tx.Tag.WithContext(ctx).Where(s.tx.Tag.ID.Eq(id)).
		UpdateColumns(map[string]interface{}{
			"pull_times": gorm.Expr("pull_times + ?", 1),
			"last_pull":  time.Now(),
		})
	if err != nil {
		return err
	}
	return nil
}

// ListByDtPagination lists the tags by the specified repository and pagination.
func (s *tagService) ListByDtPagination(ctx context.Context, repository string, limit int, lastID ...uint) ([]*models.Tag, error) {
	do := s.tx.Tag.WithContext(ctx).
		LeftJoin(s.tx.Repository, s.tx.Tag.RepositoryID.EqCol(s.tx.Repository.ID)).
		Where(s.tx.Repository.Name.Eq(repository))
	if len(lastID) > 0 {
		do = do.Where(s.tx.Tag.ID.Gt(lastID[0]))
	}
	tags, err := do.Order(s.tx.Tag.ID).Limit(limit).Find()
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// ListTag lists the tags by the specified request.
func (s *tagService) ListTag(ctx context.Context, req types.ListTagRequest) ([]*models.Tag, error) {
	return s.tx.Tag.WithContext(ctx).
		LeftJoin(s.tx.Repository, s.tx.Tag.RepositoryID.EqCol(s.tx.Repository.ID)).
		Where(s.tx.Repository.Name.Eq(req.Repository)).
		Offset(req.PageSize * (req.PageNum - 1)).Limit(req.PageSize).Find()
}

// CountArtifact counts the artifacts by the specified request.
func (s *tagService) CountTag(ctx context.Context, req types.ListTagRequest) (int64, error) {
	return s.tx.Tag.WithContext(ctx).
		LeftJoin(s.tx.Repository, s.tx.Tag.RepositoryID.EqCol(s.tx.Repository.ID)).
		Where(s.tx.Repository.Name.Eq(req.Repository)).
		Count()
}

// DeleteByID deletes the tag with the specified tag ID.
func (s *tagService) DeleteByID(ctx context.Context, id uint) error {
	matched, err := s.tx.Tag.WithContext(ctx).Where(s.tx.Tag.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
