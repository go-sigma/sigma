package types

// ArtifactItem represents an artifact.
type ArtifactItem struct {
	ID        uint   `json:"id"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ListArtifactRequest represents the request to list artifacts.
type ListArtifactRequest struct {
	Pagination

	Namespace  string `json:"namespace" param:"namespace" validate:"required,min=2,max=20,is_valid_namespace"`
	Repository string `json:"repository" query:"repository" validate:"required,is_valid_repository"`
}

// GetArtifactRequest represents the request to get an artifact.
type GetArtifactRequest struct {
	ID         uint   `json:"id" param:"id" validate:"required,number"`
	Namespace  string `json:"namespace" param:"namespace" validate:"required,min=2,max=20,is_valid_namespace"`
	Repository string `json:"repository" query:"repository" validate:"required,is_valid_repository"`
	Digest     string `json:"digest" param:"digest" validate:"required,is_valid_digest"`
}

// DeleteArtifactRequest represents the request to delete an artifact.
type DeleteArtifactRequest struct {
	ID         uint   `json:"id" param:"id" validate:"required,number"`
	Namespace  string `json:"namespace" param:"namespace" validate:"required,min=2,max=20,is_valid_namespace"`
	Repository string `json:"repository" query:"repository" validate:"required,is_valid_repository"`
	Digest     string `json:"digest" param:"digest" validate:"required,is_valid_digest"`
}
