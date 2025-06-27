package cmd

import (
	"context"
	"http-diff/cmd/task"
	"http-diff/lib/config"
	"http-diff/lib/http"
	"http-diff/lib/logger"
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

		config.Init(configFile, cfg)

		logger.Init("Http-Diff", cfg.LoggerConfig)

		http.Init(cfg.FastHttp)

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		logger.Info(ctx, "http-diff started")

		dispatcher, err := task.NewDispatcher(ctx, cfg.DiffConfigs)
		if err != nil {
			logger.Error(ctx, "failed to create task dispatcher", zap.Error(err))
			return err
		}

		// 启动任务
		go dispatcher.Start()

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
