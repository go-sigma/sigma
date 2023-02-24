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

	Namespace string `json:"namespace" param:"namespace"`
}

// GetRepositoryRequest represents the request to get a repository.
type GetRepositoryRequest struct {
	ID        uint   `json:"name" param:"id" validate:"required,number"`
	Namespace string `json:"namespace" param:"namespace" validate:"required,min=2,max=20"`
}
