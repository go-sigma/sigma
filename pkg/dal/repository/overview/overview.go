// Copyright 2026 sigma
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package overview

import (
	"context"
	"time"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
)

//go:generate go tool mockgen -destination=mocks/overview_mocks.go -package=mocks github.com/go-sigma/sigma/pkg/dal/repository/overview OverviewRepository

// OverviewRepository aggregates the cross domain queries of the home dashboard.
type OverviewRepository interface {
	// ListManagedNamespaceIDs returns the namespaces the user can manage, which are the
	// namespaces the user is a manager or an admin of.
	ListManagedNamespaceIDs(ctx context.Context, userID string) ([]string, error)
	// AggregateNamespaces sums the usage of the given namespaces.
	// A nil slice means every namespace of the instance.
	AggregateNamespaces(ctx context.Context, namespaceIDs []string) (*api.NamespaceUsage, error)
	// TopNamespaces lists namespaces ordered by storage usage.
	// A nil slice means every namespace of the instance.
	TopNamespaces(ctx context.Context, namespaceIDs []string, limit int) ([]*models.Namespace, error)
	// SumBlobSize sums the distinct blobs reachable from the given namespaces. Images share
	// layers, so summing the per artifact sizes would count the same blob many times.
	// A nil slice means every namespace of the instance.
	SumBlobSize(ctx context.Context, namespaceIDs []string) (int64, error)
	// ListBuildRecords returns the builds that finished after the cutoff. An empty userID
	// means every build of the instance.
	ListBuildRecords(ctx context.Context, since time.Time, userID string) ([]BuildRecord, error)
}

// BuildRecord is the build data the dashboard aggregates per day.
type BuildRecord struct {
	StartedAt int64
	EndedAt   *int64
	Status    enums.BuildStatus
}

type overviewRepository struct {
	tx *query.Query
}

// NewOverviewRepository creates a new overview repository with the optional query transaction
func NewOverviewRepository(txs ...*query.Query) OverviewRepository {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &overviewRepository{tx: tx}
}

// ListManagedNamespaceIDs returns the namespaces the user is a manager or an admin of.
func (s *overviewRepository) ListManagedNamespaceIDs(ctx context.Context, userID string) ([]string, error) {
	ids := make([]string, 0)
	err := s.tx.UnderlyingDB().WithContext(ctx).
		Table("namespaces AS n").
		Select("DISTINCT n.id").
		Joins("JOIN namespace_members AS m ON m.namespace_id = n.id").
		Where("n.deleted_at = 0 AND m.deleted_at = 0 AND m.user_id = ?", userID).
		Where("m.role IN ?", []enums.NamespaceRole{enums.NamespaceRoleManager, enums.NamespaceRoleAdmin}).
		Scan(&ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// AggregateNamespaces sums the usage of the given namespaces.
//
// The namespace size is maintained incrementally, so it is read from the namespace row;
// repositories and tags are counted on the fly because their cached counters can lag behind
// after a restore or a manual cleanup.
func (s *overviewRepository) AggregateNamespaces(ctx context.Context, namespaceIDs []string) (*api.NamespaceUsage, error) {
	usage := &api.NamespaceUsage{}
	if namespaceIDs != nil && len(namespaceIDs) == 0 {
		return usage, nil
	}

	db := s.tx.UnderlyingDB().WithContext(ctx)

	namespaces := struct {
		Namespaces int64
		Size       int64
	}{}
	q := db.Table("namespaces").
		Select("COUNT(*) AS namespaces, COALESCE(SUM(size), 0) AS size").
		Where("deleted_at = 0")
	if namespaceIDs != nil {
		q = q.Where("id IN ?", namespaceIDs)
	}
	if err := q.Scan(&namespaces).Error; err != nil {
		return nil, err
	}
	usage.Namespaces = namespaces.Namespaces
	usage.Size = namespaces.Size

	repositoriesQuery := db.Table("repositories").Where("deleted_at = 0")
	if namespaceIDs != nil {
		repositoriesQuery = repositoriesQuery.Where("namespace_id IN ?", namespaceIDs)
	}
	if err := repositoriesQuery.Count(&usage.Repositories).Error; err != nil {
		return nil, err
	}

	tagsQuery := db.Table("tags AS t").
		Joins("JOIN repositories AS r ON r.id = t.repository_id AND r.deleted_at = 0").
		Where("t.deleted_at = 0")
	if namespaceIDs != nil {
		tagsQuery = tagsQuery.Where("r.namespace_id IN ?", namespaceIDs)
	}
	if err := tagsQuery.Count(&usage.Tags).Error; err != nil {
		return nil, err
	}

	return usage, nil
}

// TopNamespaces lists namespaces ordered by storage usage and fills their live counters.
func (s *overviewRepository) TopNamespaces(ctx context.Context, namespaceIDs []string, limit int) ([]*models.Namespace, error) {
	namespaces := make([]*models.Namespace, 0, limit)
	if namespaceIDs != nil && len(namespaceIDs) == 0 {
		return namespaces, nil
	}

	db := s.tx.UnderlyingDB().WithContext(ctx)
	q := db.Model(&models.Namespace{})
	if namespaceIDs != nil {
		q = q.Where("id IN ?", namespaceIDs)
	}
	if err := q.Order("size DESC").Limit(limit).Find(&namespaces).Error; err != nil {
		return nil, err
	}
	if len(namespaces) == 0 {
		return namespaces, nil
	}

	ids := make([]string, 0, len(namespaces))
	for _, namespaceObj := range namespaces {
		ids = append(ids, namespaceObj.ID)
	}

	repositories, err := s.countByNamespace(ctx, "repositories", ids)
	if err != nil {
		return nil, err
	}
	tags, err := s.countByNamespace(ctx, "tags", ids)
	if err != nil {
		return nil, err
	}

	for _, namespaceObj := range namespaces {
		namespaceObj.RepositoryCount = repositories[namespaceObj.ID]
		namespaceObj.TagCount = tags[namespaceObj.ID]
	}
	return namespaces, nil
}

// countByNamespace groups the live rows of repositories or tags by namespace id.
func (s *overviewRepository) countByNamespace(ctx context.Context, table string, namespaceIDs []string) (map[string]int64, error) {
	db := s.tx.UnderlyingDB().WithContext(ctx)
	counts := make(map[string]int64, len(namespaceIDs))
	rows := make([]struct {
		NamespaceID string
		Count       int64
	}, 0)

	var q = db
	if table == "tags" {
		q = q.Table("tags AS t").
			Select("r.namespace_id AS namespace_id, COUNT(*) AS count").
			Joins("JOIN repositories AS r ON r.id = t.repository_id AND r.deleted_at = 0").
			Where("r.namespace_id IN ? AND t.deleted_at = 0", namespaceIDs).
			Group("r.namespace_id")
	} else {
		q = q.Table("repositories").
			Select("namespace_id AS namespace_id, COUNT(*) AS count").
			Where("namespace_id IN ? AND deleted_at = 0", namespaceIDs).
			Group("namespace_id")
	}
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		counts[row.NamespaceID] = row.Count
	}
	return counts, nil
}

// SumBlobSize sums the distinct blobs reachable from the given namespaces.
func (s *overviewRepository) SumBlobSize(ctx context.Context, namespaceIDs []string) (int64, error) {
	if namespaceIDs != nil && len(namespaceIDs) == 0 {
		return 0, nil
	}

	db := s.tx.UnderlyingDB().WithContext(ctx)
	total := struct {
		Size int64
	}{}
	q := db.Table("blobs AS b").
		Select("COALESCE(SUM(b.size), 0) AS size").
		Where("b.deleted_at = 0")
	if namespaceIDs != nil {
		reachable := db.Table("artifact_blobs AS ab").
			Select("ab.blob_id").
			Joins("JOIN artifacts AS a ON a.id = ab.artifact_id AND a.deleted_at = 0").
			Joins("JOIN repositories AS r ON r.id = a.repository_id AND r.deleted_at = 0").
			Where("r.namespace_id IN ?", namespaceIDs)
		q = q.Where("b.id IN (?)", reachable)
	}
	if err := q.Scan(&total).Error; err != nil {
		return 0, err
	}
	return total.Size, nil
}

// ListBuildRecords returns the builds that finished after the cutoff.
//
// A build belongs to the user that triggered it; builds started by cron or webhook have no
// acting user, so they fall back to the user that configured the builder.
func (s *overviewRepository) ListBuildRecords(ctx context.Context, since time.Time, userID string) ([]BuildRecord, error) {
	records := make([]BuildRecord, 0)
	q := s.tx.UnderlyingDB().WithContext(ctx).
		Table("builder_runners AS r").
		Select("COALESCE(r.started_at, r.created_at) AS started_at, r.ended_at AS ended_at, r.status AS status").
		Joins("JOIN builders AS b ON b.id = r.builder_id").
		Where("r.deleted_at = 0 AND b.deleted_at = 0").
		Where("COALESCE(r.started_at, r.created_at) >= ?", since.UTC().UnixMilli())
	if userID != "" {
		q = q.Where("COALESCE(r.triggered_by, b.creator_id) = ?", userID)
	}
	if err := q.Scan(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}
