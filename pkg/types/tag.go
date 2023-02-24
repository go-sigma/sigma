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
}
