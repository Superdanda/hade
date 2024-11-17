package base

// PageRequest 分页请求结构体
type PageRequest struct {
	PageNumber int `json:"page_number"` // 当前页码，默认是第一页
	PageSize   int `json:"page_size"`   // 每页记录数，默认是10条
}
