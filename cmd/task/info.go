package task

// Info 任务信息
type Info struct {
	Method      string `json:"method"`      //请求方法 GET、POST
	Url         string `json:"url"`         // 请求地址
	ContentType string `json:"contentType"` // 请求内容类型
}
