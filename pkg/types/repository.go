package types

// RepositoryItem represents a repository.
type RepositoryItem struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`

	ArtifactCount int64 `json:"artifact_count"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ListRepositoryRequest represents the request to list repositories.
type ListRepositoryRequest struct {
	Pagination
}
