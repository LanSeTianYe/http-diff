package cmd

import (
	"context"
	"fmt"

	"http-diff/cmd/task"
	"http-diff/lib/config"
	"http-diff/lib/http"
	"http-diff/lib/logger"
	"http-diff/lib/safe"
	"http-diff/lib/signal"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var configFile = ""
var cfg = &config.Configs{}

func init() {
	initStartFlag()
	rootCmd.AddCommand(startCmd)
}

func initStartFlag() {
	flags := startCmd.PersistentFlags()

	flags.StringVarP(&configFile, "config", "c", "./config/config.toml", "配置文件")
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "开始运行数据对比任务",
	Long:  "开始运行数据对比任务",
	RunE: func(cmd *cobra.Command, args []string) error {

		if err := config.Init(configFile, cfg); err != nil {
			return err
		}

		logger.Init("Http-Diff", cfg.LoggerConfig)

		http.Init(cfg.FastHttp)

		if validatedDiffConfig, err := validateDiffConfig(cfg.DiffConfigs); err != nil {
			return err
		} else {
			cfg.DiffConfigs = validatedDiffConfig
		}

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		logger.Info(ctx, "http-diff started")

		dispatcher, err := task.NewDispatcher(ctx, cfg.DiffConfigs)
		if err != nil {
			logger.Error(ctx, "failed to create task dispatcher", zap.Error(err))
			return err
		}

		// 启动任务
		go safe.RecoveryWithLogger(dispatcher.Start, ctx, "Dispatcher_Start")

		//等待程序运行结束或者接收到终止信号
		select {
		case <-dispatcher.Done():
			logger.Info(ctx, "http-diff stopped, all tasks completed")
		case sig := <-signal.ReceiveShutdownSignal():
			logger.Info(ctx, "http-diff received shutdown signal", zap.String("signal", sig.String()))
		}

		return nil
	},
}

// 验证参数
func validateDiffConfig(diffConfigs []config.DiffConfig) ([]config.DiffConfig, error) {
	if len(diffConfigs) == 0 {
		return nil, fmt.Errorf("diffConfig cannot be empty")
	}

	result := make([]config.DiffConfig, 0, len(diffConfigs))
	taskNameIndexMap := make(map[string]int)

	for index, diffConfig := range diffConfigs {

		if diffConfig.Name == "" {
			return nil, fmt.Errorf("diff config name cannot be empty,index:[%d], config detial:[%v]", index, diffConfig)
		}

		// 检查任务名称是否重复
		if existingIndex, exists := taskNameIndexMap[diffConfig.Name]; exists {
			return nil, fmt.Errorf("diff config name is duplicated: %s, first defined at index [%d], current index [%d]", diffConfig.Name, existingIndex, index)
		}

		taskNameIndexMap[diffConfig.Name] = index

		// 设置默认值
		if diffConfig.Concurrency <= 0 {
			diffConfig.Concurrency = 1
		}

		if diffConfig.WorkDir == "" {
			return nil, fmt.Errorf("diff config work_dir cannot be empty,index:[%d], config detial:[%v]", index, diffConfig)
		}

		if diffConfig.Payload == "" {
			return nil, fmt.Errorf("diff config payload cannot be empty,index:[%d], config detial:[%v]", index, diffConfig)
		}

		if diffConfig.UrlA == "" {
			return nil, fmt.Errorf("diff config url_a cannot be empty,index:[%d], config detial:[%v]", index, diffConfig)
		}

		if diffConfig.UrlB == "" {
			return nil, fmt.Errorf("diff config url_b cannot be empty,index:[%d], config detial:[%v]", index, diffConfig)
		}

		if diffConfig.Method == "" {
			return nil, fmt.Errorf("diff config method cannot be empty,index:[%d], config detial:[%v]", index, diffConfig)
		}

		result = append(result, diffConfig)
	}

	return result, nil
}
