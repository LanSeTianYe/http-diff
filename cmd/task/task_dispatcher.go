package task

import (
	"context"
	"strings"
	"sync"

	"http-diff/lib/config"
	"http-diff/lib/logger"

	"go.uber.org/zap"
)

// Dispatcher 调度器
type Dispatcher struct {

	// ctx 上下文，用于控制任务的生命周期
	ctx context.Context

	// tasks 存储所有的任务
	tasks []*Task

	// waitGroup 用于等待所有任务完成
	waitGroup *sync.WaitGroup

	// done channel 用于通知任务完成
	done chan struct{}
}

func NewDispatcher(ctx context.Context, diffConfigs []config.DiffConfig) (*Dispatcher, error) {

	dispatcher := &Dispatcher{
		ctx:       context.Background(),
		tasks:     make([]*Task, 0, len(diffConfigs)),
		waitGroup: &sync.WaitGroup{},
		done:      make(chan struct{}),
	}

	for _, diffConfig := range diffConfigs {

		taskConfig := initTaskConfig(diffConfig)

		// 初始化任务
		task, err := InitTask(ctx, taskConfig)
		if err != nil {
			logger.Error(ctx, "NewDispatcher failed to initialize task", zap.Any("config", taskConfig), zap.String("taskName", diffConfig.Name), zap.Error(err))
			return nil, err
		}

		dispatcher.tasks = append(dispatcher.tasks, task)
	}

	return dispatcher, nil
}

func (d *Dispatcher) Start() {

	logger.Info(d.ctx, "Dispatcher started", zap.Any("dispatcher", d))

	for _, task := range d.tasks {
		d.waitGroup.Add(1)
		go func(taskParam *Task) {
			taskParam.Run()
			d.waitGroup.Done()
		}(task)
	}

	//等待任务结束
	d.waitGroup.Wait()

	logger.Info(d.ctx, "Dispatcher stopped", zap.Any("dispatcher", d))

	close(d.done)
}

func (d *Dispatcher) Done() <-chan struct{} {
	return d.done
}

// initTaskConfig 初始化任务配置
func initTaskConfig(diffConfig config.DiffConfig) Config {
	return Config{
		TaskName:             diffConfig.Name,
		WorkDir:              diffConfig.WorkDir,
		Payload:              diffConfig.Payload,
		WaitTime:             diffConfig.WaitTime,
		Concurrency:          diffConfig.Concurrency,
		UrlA:                 diffConfig.UrlA,
		UrlB:                 diffConfig.UrlB,
		Method:               diffConfig.Method,
		ContentType:          diffConfig.ContentType,
		IgnoreFields:         strings.Split(diffConfig.IgnoreFields, ","),
		OutputShowNoDiffLine: diffConfig.OutputShowNoDiffLine,
		LogStatistics:        diffConfig.LogStatistics,
		SuccessConditions:    diffConfig.SuccessConditions,
	}
}
