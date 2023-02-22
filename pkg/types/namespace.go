package types

// Namespace represents a namespace.
type Namespace struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name" validate:"required,min=2,max=20"`
	Description *string `json:"description" validate:"max=30"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}
