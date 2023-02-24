package types

// ArtifactItem represents an artifact.
type ArtifactItem struct {
	ID        uint   `json:"id"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
