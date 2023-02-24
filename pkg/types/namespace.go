package types

// NamespaceItem represents a namespace.
type NamespaceItem struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name" validate:"required,min=2,max=20,is_valid_namespace"`
	Description *string `json:"description" validate:"max=30"`

	ArtifactCount int64 `json:"artifact_count"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// CreateNamespaceRequest represents the request to create a namespace.
type CreateNamespaceRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=20,is_valid_namespace"`
	Description *string `json:"description" validate:"max=30"`
}

// ListNamespaceRequest represents the request to list namespaces.
type ListNamespaceRequest struct {
	Pagination

	// Name query the namespace by name.
	Name *string `json:"name" query:"name" validate:"min=2,max=20,is_valid_namespace"`
}

// GetNamespaceRequest represents the request to get a namespace.
type GetNamespaceRequest struct {
	ID uint `json:"id" param:"id" validate:"required,number"`
}

// DeleteNamespaceRequest represents the request to delete a namespace.
type DeleteNamespaceRequest struct {
	ID uint `json:"id" param:"id" validate:"required,number"`
}

// PutNamespaceRequest represents the request to update a namespace.
type PutNamespaceRequest struct {
	ID uint `json:"id" param:"id" validate:"required,number"`

	Description *string `json:"description" validate:"max=30"`

	ArtifactCount int64 `json:"artifact_count"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
