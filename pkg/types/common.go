package types

// CommonList is the common list struct
type CommonList struct {
	Total int64 `json:"total"`
	Items []any `json:"items"`
}

// Pagination is the pagination struct
type Pagination struct {
	PageSize int `json:"page_size" query:"page_size" validate:"required,gte=10,lte=100"`
	PageNum  int `json:"page_num" query:"page_num" validate:"required,gte=1"`
}
