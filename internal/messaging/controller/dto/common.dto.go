package dto

type Meta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

type PaginatedResponse struct {
	Data interface{} `json:"data"`
	Meta Meta        `json:"meta"`
}

func NewPaginatedResponse(data interface{}, page, limit int, total int64) *PaginatedResponse {
	return &PaginatedResponse{
		Data: data,
		Meta: Meta{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	}
}
