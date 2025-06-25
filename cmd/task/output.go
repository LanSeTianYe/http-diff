package task

type OutPut struct {
	Payload *Payload `json:"payload"` // 请求负载

	UrlAResponse interface{} `json:"urlAResponse"` // urlA 响应
	UrlBResponse interface{} `json:"urlBResponse"` // urlB 响应

	Diff string `json:"diff"` //响应对比结果
}

// FailedOutPut 出错时的信息
type FailedOutPut struct {
	Params  string `json:"params"`
	Headers string `json:"headers"`
	Body    string `json:"body"`
	Err     string `json:"err"`
}

func NewFailedOuyPut(payload *Payload, err error) *FailedOutPut {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}

	return &FailedOutPut{
		Params:  payload.Params,
		Headers: payload.Headers,
		Body:    payload.Body,
		Err:     errStr,
	}
}
