package task

import (
	"bufio"
	"context"
	"errors"
	"github.com/spf13/cast"
	"http-diff/lib/concurrency"
	"http-diff/lib/safe"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"http-diff/lib/logger"
	"http-diff/util"

	"github.com/bytedance/sonic"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
)

type Task struct {
	ctx context.Context
	// stopCh 用于通知任务停止
	stopCh chan struct{}
	// stopChOnce 用于确保 stopCh 只被关闭一次
	stopChOnce *sync.Once

	// Config 任务配置
	Config Config
	// SuccessConditionMap 接口响应成功的条件
	SuccessConditionMap map[string]string

	// waitGroup 用户等待任务的子程序结束
	waitGroup *sync.WaitGroup
	// statisticsInfo 任务统计信息
	statisticsInfo *StatisticsInfo

	// UrlAInfo 接口A请求信息
	UrlAInfo *Info
	// UrlBInfo 接口B请求信息
	UrlBInfo *Info

	// inputCh 输入通道，用于接收待处理的 Payload
	inputCh chan *Payload
	// outputCh 输出通道，用于发送处理结果
	outputCh chan *OutPut
	// failedCH 错误输出通道，用于发送处理错误信息
	failedCH chan *FailedOutPut
}

type Config struct {
	// TaskName 任务名称
	TaskName string
	// WorkDir 工作目录
	WorkDir string
	// Payload 文件路径或内容
	Payload string
	// WaitTime 等待时间
	WaitTime time.Duration

	// Concurrency 并发数
	Concurrency int
	// UrlA 接口A地址
	UrlA string
	// UrlB 接口B地址
	UrlB string
	// Method HTTP方法
	Method string
	// ContentType 内容类型
	ContentType string
	// IgnoreFields 忽略的字段
	IgnoreFields []string
	// OutputShowNoDiffLine 是否输出没有差异的行
	OutputShowNoDiffLine bool
	// LogStatistics 是在日志中录统计信息
	LogStatistics bool
	// SuccessConditions 接口响应成功的条件，避免调用接口返回错误相同的错误码，但是diff是空的情况
	SuccessConditions string
}

func InitTask(ctx context.Context, cfg Config) (*Task, error) {

	files := strings.Split(cfg.Payload, ",")
	lineCount := 0
	for _, file := range files {
		filePath := path.Join(cfg.WorkDir, file)

		count, err := util.FileLineCount(filePath)
		if err != nil {
			return nil, err
		}

		lineCount = lineCount + count
	}

	task := &Task{
		ctx:        ctx,
		stopCh:     make(chan struct{}),
		stopChOnce: &sync.Once{},

		Config:              cfg,
		SuccessConditionMap: make(map[string]string),
		waitGroup:           &sync.WaitGroup{},
		statisticsInfo:      NewStatisticsInfo(lineCount),
		UrlAInfo: &Info{
			Method:      cfg.Method,
			Url:         cfg.UrlA,
			ContentType: cfg.ContentType,
		},
		UrlBInfo: &Info{
			Method:      cfg.Method,
			Url:         cfg.UrlB,
			ContentType: cfg.ContentType,
		},
		inputCh:  make(chan *Payload, 1000),
		outputCh: make(chan *OutPut, 1000),
		failedCH: make(chan *FailedOutPut, 1000),
	}

	if cfg.SuccessConditions != "" {
		conditions := strings.Split(cfg.SuccessConditions, ",")
		for _, condition := range conditions {
			parts := strings.SplitN(condition, "=", 2)
			if len(parts) == 2 {
				task.SuccessConditionMap[parts[0]] = parts[1]
			} else {
				logger.Error(ctx, "InitTask Invalid success condition format", zap.String("condition", condition))
				return nil, errors.New("invalid success condition format: " + condition)
			}
		}
	}

	return task, nil
}

func (t *Task) Run() {
	logger.Info(t.ctx, "Task_Run Start running task", zap.Any("task", t))

	// 读文件
	go safe.RecoveryWithLoggerAndCallback(t.runReader, t.ctx, "Task_Run_runReader", func() { t.stop() })
	time.Sleep(time.Second * 2) // 等待文件读取完成，避免在文件读取过程中就开始处理请求

	// 处理请求
	for i := 0; i < t.Config.Concurrency; i++ {
		go safe.RecoveryWithLoggerAndCallback(t.run, t.ctx, "Task_Run_run", func() { t.stop() })
	}

	// 写结果
	go safe.RecoveryWithLoggerAndCallback(func() {
		if err := t.writeOutputToFile(); err != nil {
			t.stop()
		}
	}, t.ctx, "Task_Run_writeOutputToFile", func() { t.stop() })

	// 写错误数据
	go safe.RecoveryWithLoggerAndCallback(func() {
		if err := t.writeFailedPayloadToFile(); err != nil {
			t.stop()
		}
	}, t.ctx, "Task_Run_writeFailedPayloadToFile", func() { t.stop() })

	// 记录统计信息
	if t.Config.LogStatistics {
		go safe.RecoveryWithLoggerAndCallback(t.logStatisticsInfoLoop, t.ctx, "Task_Run_logStatisticsInfoLoop", func() { t.stop() })
	}

	// 等代任务运行完成
	go func() {
		t.waitGroup.Wait()
		t.stop()
	}()

	// 等待任务结束信号
	<-t.Done()

	//任务运行结束的时候打印一次日志
	if t.Config.LogStatistics {
		t.logStatisticsInfo()
	}

	logger.Info(t.ctx, "Task_Run Stop running task", zap.Any("task", t))
}

func (t *Task) runReader() {
	payLoadFiles := strings.Split(t.Config.Payload, ",")

	logger.Info(t.ctx, "Task_runReader Starting to read payload files", zap.Strings("files", payLoadFiles))

	for _, payLoadFile := range payLoadFiles {
		filePath := path.Join(t.Config.WorkDir, payLoadFile)
		logger.Warn(t.ctx, "Task_runReader Reading file start", zap.String("file", filePath))

		file, err := os.Open(filePath)
		if err != nil {
			logger.Error(t.ctx, "Task_runReader Failed to open file", zap.String("filePath", filePath), zap.Error(err))
			panic(err)
		}

		maxLineSize := 1024 * 1024
		buffer := make([]byte, 0, maxLineSize)
		scanner := bufio.NewScanner(file)
		scanner.Buffer(buffer, maxLineSize)
		lineNumber := 0

		for scanner.Scan() {
			line := scanner.Text()
			lineNumber++

			logger.Debug(t.ctx, "Task_runReader Read line from file", zap.String("line", line), zap.Int("lineNumber", lineNumber))

			if len(line) == 0 {
				t.statisticsInfo.AddFailed()
				logger.Error(t.ctx, "Task_runReader Empty line in file", zap.String("filePath", filePath), zap.Int("lineNumber", lineNumber))
				continue
			}

			if err := scanner.Err(); err != nil {
				t.statisticsInfo.AddFailed()
				logger.Error(t.ctx, "Task_runReader Error reading file", zap.String("filePath", filePath), zap.Int("lineNumber", lineNumber), zap.Error(err))
				continue
			}

			payload := &Payload{}
			err := sonic.Unmarshal([]byte(line), payload)
			if err != nil {
				t.statisticsInfo.AddFailed()
				logger.Error(t.ctx, "Task_runReader Failed to unmarshal payload", zap.String("line", line), zap.Int("lineNumber", lineNumber), zap.Error(err))
				continue
			}

			t.waitGroup.Add(1)

			logger.Debug(t.ctx, "Task_runReader Adding payload to input channel", zap.Any("payload", payload))
			t.inputCh <- payload
		}

		logger.Warn(t.ctx, "Task_runReader Reading file end", zap.String("file", filePath))
	}

	logger.Info(t.ctx, "Task_runReader Finished reading all payload files", zap.Any("files", payLoadFiles))
}

func (t *Task) run() {
	for {
	SelectLoop:
		select {
		case <-t.ctx.Done():
			return
		case payload := <-t.inputCh:
			logger.Debug(t.ctx, "Task_run Processing payload", zap.String("task", t.Config.TaskName), zap.Any("payload", payload))

			if t.Config.WaitTime != time.Duration(0) {
				logger.Debug(t.ctx, "Task_run Waiting for specified time", zap.Duration("waitTime", t.Config.WaitTime))
				time.Sleep(t.Config.WaitTime)
			}

			var urlAResponse interface{}
			var urlAResponseErr error
			var urlBResponse interface{}
			var urlBResponseErr error

			safeGoWaitGroup := concurrency.NewSafeGoWaitGroup()
			safeGoWaitGroup.SafeGoWithLogger(func() {
				urlAResponse, urlAResponseErr = DoRequest(t.ctx, t.UrlAInfo, payload)
			}, func(message any) {
				logger.Error(t.ctx, "Task_run Failed to get response from UrlA", zap.Any("urlA", t.UrlAInfo), zap.Any("payload", payload), zap.Any("message", message))
				urlAResponseErr = errors.New("failed to get response from UrlA: " + cast.ToString(message))
			})

			safeGoWaitGroup.SafeGoWithLogger(func() {
				urlBResponse, urlBResponseErr = DoRequest(t.ctx, t.UrlBInfo, payload)
			}, func(message any) {
				logger.Error(t.ctx, "Task_run Failed to get response from UrlB", zap.Any("urlB", t.UrlBInfo), zap.Any("payload", payload), zap.Any("message", message))
				urlBResponseErr = errors.New("failed to get response from UrlB: " + cast.ToString(message))
			})

			safeGoWaitGroup.Wait()

			if urlAResponseErr != nil || urlBResponseErr != nil {
				logger.Error(t.ctx, "Task_run Failed to get response", zap.Any("payload", payload), zap.Any("urlAResponseErr", urlAResponseErr), zap.Any("urlBResponseErr", urlBResponseErr))
				t.statisticsInfo.AddFailed()
				t.failedCH <- NewFailedOuyPut(payload, errors.New("failed to get response: "+cast.ToString(urlAResponseErr)+"; "+cast.ToString(urlBResponseErr)))
				break SelectLoop
			}

			if !t.responseSuccess(urlAResponse) || !t.responseSuccess(urlBResponse) {
				logger.Error(t.ctx, "Task_run Response does not meet success conditions", zap.Any("payload", payload), zap.Any("urlAResponse", urlAResponse), zap.Any("urlBResponse", urlBResponse))
				t.failedCH <- NewFailedOuyPut(payload, errors.New("response does not meet success conditions"))
				t.statisticsInfo.AddFailed()
				break SelectLoop
			}

			urlAResponseFieldMap := make(map[string]interface{})
			urlBResponseFieldMap := make(map[string]interface{})

			if len(t.Config.IgnoreFields) != 0 {
				for _, field := range t.Config.IgnoreFields {
					urlAValue, err := util.SetJsonFieldToNil(urlAResponse, field)
					if err != nil {
						logger.Error(t.ctx, "Task_run Failed to set field to nil in urlA response", zap.Any("response", urlAResponse), zap.Any("field", field), zap.Error(err))
						t.failedCH <- NewFailedOuyPut(payload, err)
						t.statisticsInfo.AddFailed()
						break SelectLoop
					}
					urlAResponseFieldMap[field] = urlAValue

					urlBValue, err := util.SetJsonFieldToNil(urlBResponse, field)
					if err != nil {
						t.failedCH <- NewFailedOuyPut(payload, err)
						t.statisticsInfo.AddFailed()
						logger.Error(t.ctx, "Task_run Failed to set field to nil in urlB response", zap.Any("response", urlBResponse), zap.Any("field", field), zap.Error(err))
						break SelectLoop
					}
					urlBResponseFieldMap[field] = urlBValue
				}
			}

			diff := cmp.Diff(urlAResponse, urlBResponse)
			if diff == "" {
				t.outputCh <- &OutPut{Payload: payload, Diff: diff, UrlAResponse: nil, UrlBResponse: nil}
				t.statisticsInfo.AddSame()
				break SelectLoop
			}

			urlAErr := t.recoverFileValue(urlAResponse, urlAResponseFieldMap)
			urlBErr := t.recoverFileValue(urlBResponse, urlBResponseFieldMap)
			if urlAErr != nil || urlBErr != nil {
				logger.Error(t.ctx, "Task_run Failed to set field value in response", zap.Any("payload", payload), zap.Any("urlAErr", urlAErr), zap.Any("urlBErr", urlBErr))
				t.failedCH <- NewFailedOuyPut(payload, errors.New("failed to set field value in response: "+cast.ToString(urlAErr)+"; "+cast.ToString(urlBErr)))
				t.statisticsInfo.AddFailed()
				break SelectLoop
			}

			t.statisticsInfo.AddDiff()
			t.outputCh <- &OutPut{Payload: payload, Diff: diff, UrlAResponse: urlAResponse, UrlBResponse: urlBResponse}
		}
	}
}

func (t *Task) recoverFileValue(jsonData interface{}, valueMap map[string]interface{}) error {
	for key, value := range valueMap {
		err := util.SetJsonFieldValue(jsonData, key, value)
		if err != nil {
			logger.Error(t.ctx, "Task_recoverFileValue Failed to set field value in jsonData", zap.Any("jsonData", jsonData), zap.Any("field", key), zap.Any("value", value), zap.Error(err))
			return err
		}
	}
	return nil
}

// writeOutputToFile 用于将输出结果写入文件
func (t *Task) writeOutputToFile() error {

	outputFilePath := path.Join(t.Config.WorkDir, t.Config.TaskName+"_output.txt")

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		logger.Error(t.ctx, "Task_writeOutputToFile Failed to create output file", zap.String("outputFilePath", outputFilePath), zap.Error(err))
		return err
	}

	defer outputFile.Close()
	defer outputFile.Sync()

	for {
		select {
		case output := <-t.outputCh:
			if !t.Config.OutputShowNoDiffLine && output != nil && output.Diff == "" {
				logger.Debug(t.ctx, "Task_writeOutputToFile Skipping output with no diff", zap.Any("task", t), zap.Any("output", output))
				t.waitGroup.Done()
				continue
			}

			marshal, err := sonic.Marshal(output)
			if err != nil {
				logger.Error(t.ctx, "Task_writeOutputToFile Failed to marshal output", zap.Any("task", t), zap.Any("output", output), zap.Error(err))
				t.waitGroup.Done()
				continue
			}

			_, err = outputFile.WriteString(string(marshal) + "\n")
			if err != nil {
				logger.Error(t.ctx, "Task_writeOutputToFile Failed to write output to file", zap.Any("task", t), zap.Any("output", output), zap.Error(err))
				t.waitGroup.Done()
				continue
			}

			t.waitGroup.Done()
		case <-t.ctx.Done():
			logger.Debug(t.ctx, "Task_writeOutputToFile Context done, stopping writer", zap.Any("task", t))
			return nil
		}
	}
}

// writeFailedPayloadToFile 用于将处理失败的 Payload 写入文件
func (t *Task) writeFailedPayloadToFile() error {

	outputFilePath := path.Join(t.Config.WorkDir, t.Config.TaskName+"_failed_payload.txt")
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		logger.Error(t.ctx, "Task_writeFailedPayloadToFile Failed to create output file", zap.String("outputFilePath", outputFilePath), zap.Error(err))
		return err
	}

	defer outputFile.Close()
	defer outputFile.Sync()

	for {
		select {
		case output := <-t.failedCH:
			marshal, err := sonic.Marshal(output)
			if err != nil {
				logger.Error(t.ctx, "Task_writeFailedPayloadToFile Failed to marshal output", zap.Any("task", t), zap.Any("output", output), zap.Error(err))
				t.waitGroup.Done()
				continue
			}

			_, err = outputFile.WriteString(string(marshal) + "\n")
			if err != nil {
				logger.Error(t.ctx, "Task_writeFailedPayloadToFile Failed to write output to file", zap.Any("task", t), zap.Any("output", output), zap.Error(err))
				t.waitGroup.Done()
				continue
			}

			t.waitGroup.Done()
		case <-t.ctx.Done():
			logger.Debug(t.ctx, "Task_writeFailedPayloadToFile Context done, stopping writer", zap.Any("task", t))
			return nil
		}
	}
}

// logStatisticsInfo 用于记录任务的统计信息
func (t *Task) logStatisticsInfoLoop() {
	tick := time.Tick(time.Second)
	for {
		select {
		case <-t.ctx.Done():
			return
		case <-tick:
			t.logStatisticsInfo()
		}
	}
}

func (t *Task) logStatisticsInfo() {
	logger.Info(t.ctx, "Task_logStatisticsInfo_"+t.Config.TaskName+":",
		zap.Int64("totalCount:", t.statisticsInfo.GetTotalCount()),
		zap.Int64("sameCount:", t.statisticsInfo.GetSameCount()),
		zap.Int64("diffCount", t.statisticsInfo.GetDiffCount()),
		zap.Int64("failedCount:", t.statisticsInfo.GetFailedCount()),
		zap.Float64("progress", float64(t.statisticsInfo.GetFailedCount()+t.statisticsInfo.GetSameCount()+t.statisticsInfo.GetDiffCount())/float64(t.statisticsInfo.GetTotalCount())*100))
}

func (t *Task) responseSuccess(result interface{}) bool {
	for key, value := range t.SuccessConditionMap {

		fieldValue, err := util.GetFieldValue(result, key)
		if err != nil {
			logger.Debug(t.ctx, "Task_responseSuccess JsonPath lookup failed", zap.String("key", key), zap.Any("result", result), zap.Error(err))
			return false
		}

		if cast.ToString(fieldValue) != value {
			return false
		}
	}

	return true
}

func (t *Task) stop() {
	t.stopChOnce.Do(func() {
		close(t.stopCh)
	})
}

func (t *Task) Done() <-chan struct{} {
	return t.stopCh
}
