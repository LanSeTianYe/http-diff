package task

import (
	"context"
	"errors"
	"net/url"

	"http-diff/constant"
	"http-diff/lib/http"
	"http-diff/lib/logger"

	"github.com/bytedance/sonic"
	"go.uber.org/zap"
)

// DoRequest 发送请求
func DoRequest(ctx context.Context, taskInfo *Info, payload *Payload) (interface{}, error) {
	logger.Debug(ctx, "DoRequest start, request info", zap.Any("taskInfo", taskInfo), zap.Any("payload", payload))

	parseUrl, err := url.Parse(taskInfo.Url)
	if err != nil {
		logger.Error(ctx, "DoRequest url.Parse error", zap.Any("taskInfo", taskInfo), zap.Any("payload", payload), zap.Error(err))
		return nil, err
	}

	if payload.Params != "" {
		queryUnescape, err := url.QueryUnescape(payload.Params)
		if err != nil {
			logger.Error(ctx, "DoRequest url.QueryUnescape error", zap.Any("payload.Params", payload.Params), zap.Error(err))
			return nil, err
		}

		parseQuery, err := url.ParseQuery(queryUnescape)
		if err != nil {
			logger.Error(ctx, "DoRequest url.ParseQuery error", zap.Any("payload.Params", payload.Params), zap.Error(err))
			return nil, err
		}

		query := parseUrl.Query()
		for key, value := range parseQuery {
			for _, subValue := range value {
				query.Add(key, subValue)
			}
		}

		parseUrl.RawQuery = query.Encode()
	}

	requestUrl := parseUrl.String()

	logger.Debug(ctx, "DoRequest requestUrl", zap.String("url", requestUrl))

	header, err := initHeader(taskInfo, payload)
	if err != nil {
		logger.Error(ctx, "DoRequest initHeader error", zap.Any("taskInfo", taskInfo), zap.Any("payload", payload), zap.Error(err))
		return nil, err
	}
	logger.Debug(ctx, "DoRequest header", zap.Any("header", header))

	// 处理 GET 请求
	var result interface{}
	if taskInfo.Method == constant.GET {
		err := http.Get(ctx, requestUrl, nil, header, &result)
		if err != nil {
			logger.Error(ctx, "DoRequest http.Get error", zap.String("url", requestUrl), zap.Any("header", header), zap.Error(err))
			return nil, err
		}
		return result, nil
	}

	// 处理 POST 请求
	if taskInfo.Method == constant.POST {
		params, err := initPostParams(taskInfo, payload)
		if err != nil {
			logger.Error(ctx, "DoRequest initPostParams error", zap.Any("taskInfo", taskInfo), zap.Any("payload", payload), zap.Error(err))
			return nil, err
		}

		err = http.Post(ctx, requestUrl, params, header, &result)
		if err != nil {
			logger.Error(ctx, "DoRequest http.Post error", zap.String("url", requestUrl), zap.Any("params", params), zap.Any("header", header), zap.Error(err))
			return nil, err
		}

		return result, nil
	}

	// 位置类型请求
	return nil, errors.New("unsupported method: " + taskInfo.Method)
}

func initHeader(taskInfo *Info, payload *Payload) (map[string]string, error) {
	result := make(map[string]string)

	if taskInfo.ContentType != "" {
		result[constant.HeaderKeyContextType] = taskInfo.ContentType
	}

	if payload.Headers != "" {
		err := sonic.Unmarshal([]byte(payload.Headers), &result)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func initPostParams(taskInfo *Info, payload *Payload) (interface{}, error) {
	var params interface{}
	if payload.Body == "" {
		return params, nil
	}

	if taskInfo.ContentType == constant.ContentTypeForm {
		params = payload.Body
	} else {
		err := sonic.Unmarshal([]byte(payload.Body), &params)
		if err != nil {
			return nil, err
		}
	}

	return params, nil
}
