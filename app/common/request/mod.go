package request

// ModSearchRequest 搜索mod请求结构
type ModSearchRequest struct {
	Keyword    string `form:"keyword" json:"keyword"`                             // 搜索关键词
	GameID     uint   `form:"game_id" json:"game_id"`                             // 游戏ID
	CategoryID uint   `form:"category_id" json:"category_id"`                     // 分类ID
	Author     string `form:"author" json:"author"`                               // 作者
	SortBy     string `form:"sort_by" json:"sort_by"`                             // 排序字段: rating, download_count, created_at
	Order      string `form:"order" json:"order"`                                 // 排序方向: asc, desc
	Page       int    `form:"page" json:"page" binding:"min=0"`                   // 页码，允许0（控制器设置默认值）
	PageSize   int    `form:"page_size" json:"page_size" binding:"min=0,max=100"` // 页面大小，允许0（控制器设置默认值）
}

// ModDetailRequest 获取mod详情请求
type ModDetailRequest struct {
	ID uint `uri:"id" binding:"required,min=1"` // mod ID
}
