package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"http-diff/constant"
	"http-diff/lib/config"
	"http-diff/lib/logger"

	"github.com/bytedance/sonic"
	"github.com/valyala/fasthttp"
	"github.com/xiaotianfork/go-querystring-json/query"
	"go.uber.org/zap"
)

var client *fasthttp.Client
var clientOnce sync.Once

type Entity struct {
	Name string
	Id   int
}

// Init 初始化客户端配置
func Init(config config.FastHttp) {
	clientOnce.Do(func() {
		client = &fasthttp.Client{
			MaxIdemponentCallAttempts:     config.RetryTimes,
			ReadTimeout:                   config.ReadTimeOut,
			WriteTimeout:                  config.ReadTimeOut,
			MaxIdleConnDuration:           config.MaxIdleConnDuration,
			MaxConnsPerHost:               config.MaxConnsPerHost,
			NoDefaultUserAgentHeader:      false, // default User-Agent: fasthttp
			DisableHeaderNamesNormalizing: true,  // If you set the case on your headers correctly you can enable this
			DisablePathNormalizing:        true,
			MaxResponseBodySize:           10 * 1024 * 1024,
			RetryIfErr: func(request *fasthttp.Request, attempts int, err error) (resetTimeout bool, retry bool) {
				//幂等方法
				methodNeedRetry := request.Header.IsGet() || request.Header.IsHead() || request.Header.IsPut()
				if methodNeedRetry {
					return true, true
				}
				return false, false
			},
			// increase DNS cache time to an hour instead of default minute
			Dial: (&fasthttp.TCPDialer{
				Concurrency:      4096,
				DNSCacheDuration: time.Minute * 10,
			}).Dial,
		}
	})
}

func simpleGet(url string) error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodGet)

	//请求数据
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	return client.Do(req, resp)
}

// Get Get请求，不需要设置超时时间
//
// params: 请求参数，为结构体类型需要在字段后面加 json tag，会自动转换为对应的参数
//
//	type request struct {
//		Name    string   `json:"name"`
//		Age     int      `json:"age"`
//		Friends []string `json:"friends"`
//	}
func Get(ctx context.Context, requestUrl string, params interface{}, headers map[string]string, result interface{}) error {
	return GetTimeOut(ctx, requestUrl, params, headers, time.Duration(0), result)
}

// Post Post请求，不需要设置超时时间
//
// params: 请求参数，为结构体类型需要在字段后面加 json tag
//
//	type request struct {
//		Name    string   `json:"name"`
//		Age     int      `json:"age"`
//		Friends []string `json:"friends"`
//	}
func Post(ctx context.Context, requestUrl string, params interface{}, headers map[string]string, result interface{}) error {
	return PostTimeOut(ctx, requestUrl, params, headers, time.Duration(0), result)
}

// GetTimeOut Get请求，需要设置超时时间
//
// params: 请求参数，为结构体类型需要在字段后面加 json tag，会自动转换为对应的参数
//
//	type request struct {
//		Name    string   `json:"name"`
//		Age     int      `json:"age"`
//		Friends []string `json:"friends"`
//	}
func GetTimeOut(ctx context.Context, requestUrl string, params interface{}, headers map[string]string, timeOut time.Duration, result interface{}) error {
	//解析验证url
	_, err := url.Parse(requestUrl)
	if err != nil {
		return err
	}

	queryValues := url.Values{}
	if params != nil {
		queryValues, err = query.Values(params)
		if err != nil {
			return err
		}
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	if params != nil {
		req.SetRequestURI(fmt.Sprintf("%s?%s", requestUrl, queryValues.Encode()))
	} else {
		req.SetRequestURI(requestUrl)
	}

	req.Header.SetMethod(fasthttp.MethodGet)

	// 添加请求头
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	//请求数据
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err = doTimeOut(ctx, req, resp, timeOut, result)
	if err != nil {
		return err
	}

	return nil
}

// PostTimeOut Post请求，不需要设置超时时间
//
// params: 请求参数为结构体类型，需要在字段后面加 json tag，会自动转换为对应的参数
//
//	type request struct {
//		Name    string   `json:"name"`
//		Age     int      `json:"age"`
//		Friends []string `json:"friends"`
//	}
func PostTimeOut(ctx context.Context, requestUrl string, params interface{}, headers map[string]string, timeOut time.Duration, result interface{}) error {
	//解析验证url
	_, err := url.Parse(requestUrl)
	if err != nil {
		return err
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(requestUrl)

	// 添加请求头
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	// header
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType(constant.ContentTypeJson)
	if headers[constant.HeaderKeyContextType] == constant.ContentTypeForm {
		req.Header.SetContentType(constant.ContentTypeForm)
		if params != nil {
			values, err := url.ParseQuery(params.(string))
			if err != nil {
				return err
			}
			req.SetBodyString(values.Encode())
		}
	} else {
		marshal, err := sonic.Marshal(params)
		if err != nil {
			return err
		}
		req.SetBody(marshal)
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// 请求并解析数据
	err = doTimeOut(ctx, req, resp, timeOut, result)
	if err != nil {
		return err
	}

	return nil
}

func doTimeOut(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response, timeOut time.Duration, result interface{}) error {
	logger.Debug(ctx, "http_DoTimeOut", zap.Any("request", req), zap.Any("response", resp), zap.Duration("timeOut", timeOut))

	var err error

	if timeOut <= 0 {
		err = client.Do(req, resp)
	} else {
		err = client.DoTimeout(req, resp, timeOut)
	}

	if err != nil {
		return err
	}

	//状态码验证
	statusCode := resp.StatusCode()
	if statusCode != http.StatusOK {
		errInner := errors.New("data request failed , code:" + strconv.Itoa(statusCode))
		return errInner
	}

	//反序列化参数
	err = sonic.Unmarshal(resp.Body(), result)
	if err != nil {
		logger.Info(ctx, "http_DoTimeOut unmarshal error", zap.Error(err), zap.String("body", string(resp.Body())))
		return err
	}

	return nil
}
