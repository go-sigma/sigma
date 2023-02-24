package types

// NamespaceItem represents a namespace.
type NamespaceItem struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name" validate:"required,min=2,max=20"`
	Description *string `json:"description" validate:"max=30"`

	ArtifactCount int64 `json:"artifact_count"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// CreateNamespaceRequest represents the request to create a namespace.
type CreateNamespaceRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=20"`
	Description *string `json:"description" validate:"max=30"`
}

// ListNamespaceRequest represents the request to list namespaces.
type ListNamespaceRequest struct {
	Name *string `json:"name" query:"name"`
	Pagination
}
