package types

// TagItem represents an tag.
type TagItem struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ListTagRequest represents the request to list tags.
type ListTagRequest struct {
	Pagination

	Repository string `json:"repository" query:"repository" validate:"required,is_valid_repository_name"`
}

// DeleteTagRequest represents the request to delete a tag.
type DeleteTagRequest struct {
	ID         uint   `json:"id" param:"id" validate:"required,number"`
	Namespace  string `json:"namespace" param:"namespace" validate:"required,min=2,max=20,is_valid_namespace"`
	Repository string `json:"repository" query:"repository" validate:"required,is_valid_repository"`
}

// GetTagRequest represents the request to get a tag.
type GetTagRequest struct {
	ID         uint   `json:"id" param:"id" validate:"required,number"`
	Namespace  string `json:"namespace" param:"namespace" validate:"required,min=2,max=20,is_valid_namespace"`
	Repository string `json:"repository" query:"repository" validate:"required,is_valid_repository"`
}
