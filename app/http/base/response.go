package base

import "fmt"

// PageResponse 分页响应结构体
type PageResponse struct {
	TotalRecords   int64       `json:"total_records"`    // 总记录数
	TotalPages     int         `json:"total_pages"`      // 总页数
	CurrentPage    int         `json:"current_page"`     // 当前页码
	RecordsPerPage int         `json:"records_per_page"` // 每页记录数
	Data           interface{} `json:"data"`             // 当前页的数据，可以是任意类型的切片
}

// NewPageResponseWithPageRequest PageResponse构造方法
func NewPageResponseWithPageRequest(totalRecords int64, pageRequest PageRequest, data interface{}) *PageResponse {
	pr := &PageResponse{
		TotalRecords:   totalRecords,
		RecordsPerPage: pageRequest.PageSize,
		CurrentPage:    pageRequest.PageNumber,
		Data:           data,
	}
	// 计算总页数
	pr.CalculateTotalPages()
	return pr
}

// NewPageResponse PageResponse构造方法
func NewPageResponse(totalRecords int64, recordsPerPage int, currentPage int, data interface{}) *PageResponse {
	pr := &PageResponse{
		TotalRecords:   totalRecords,
		RecordsPerPage: recordsPerPage,
		CurrentPage:    currentPage,
		Data:           data,
	}
	// 计算总页数
	pr.CalculateTotalPages()
	return pr
}

// CalculateTotalPages 计算总页数
func (pr *PageResponse) CalculateTotalPages() {
	if pr.TotalRecords == 0 {
		pr.TotalPages = 0
	} else {
		pr.TotalPages = int((pr.TotalRecords + int64(pr.RecordsPerPage) - 1) / int64(pr.RecordsPerPage))
	}
}

// HasNextPage 判断当前页是否有下一页
func (pr *PageResponse) HasNextPage() bool {
	return pr.CurrentPage < pr.TotalPages
}

// HasPrevPage 判断当前页是否有上一页
func (pr *PageResponse) HasPrevPage() bool {
	return pr.CurrentPage > 1
}

// GetStartIndex 获取分页的起始索引
func (pr *PageResponse) GetStartIndex() int {
	return (pr.CurrentPage - 1) * pr.RecordsPerPage
}

// GetEndIndex 获取分页的结束索引
func (pr *PageResponse) GetEndIndex() int {
	endIndex := pr.CurrentPage * pr.RecordsPerPage
	if int64(endIndex) > pr.TotalRecords {
		endIndex = int(pr.TotalRecords)
	}
	return endIndex
}

// PrintPageInfo 打印分页信息
func (pr *PageResponse) PrintPageInfo() {
	fmt.Printf("Page %d of %d pages. Showing records %d to %d of %d total records.\n",
		pr.CurrentPage, pr.TotalPages, pr.GetStartIndex()+1, pr.GetEndIndex(), pr.TotalRecords)
}
