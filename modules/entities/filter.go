package entities

type PaginationReq struct {
	Page        int `query:"page"`
	Limit       int `query:"limit"`
	TotalPage	int `query:"total_page" json:"total_page"`
	TotalIte	int `query:"total_item" json:"total_item"`
}

type SortReq struct {
	OrderBy string `query:"order_by"`
	Sort    string `query:"sort"`  // DESC | ASC
}