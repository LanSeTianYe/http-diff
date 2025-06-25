package logger

import (
	"context"
	"testing"

	"http-diff/lib/config"
	
	"go.uber.org/zap"
)

func initLogger(name string) {
	configs := config.Configs{}
	config.Init("./data/config.toml", &configs)

	Init(name, configs.LoggerConfig)
}

func TestW(t *testing.T) {
	initLogger("TestF")

	ctx := initContext()

	Debug(ctx, "Debug")
	Debug(ctx, "Debug", zap.String("number", "1"))

	Info(ctx, "Info")
	Info(ctx, "Info", zap.String("number", "1"))

	Warn(ctx, "Warn")
	Warn(ctx, "Warn", zap.String("number", "1"))

	Error(ctx, "Error")
	Error(ctx, "Error", zap.String("number", "1"))

}

func TestMultiSingle(t *testing.T) {
	initLogger("TestMultiSingle")

	times := 1024
	for i := 0; i < times; i++ {
		Info(context.Background(), "测试打印日志", zap.String("name", "name"))
	}
}

// BenchmarkLogger-8   	  374352	      3036 ns/op
func BenchmarkLogger(b *testing.B) {
	initLogger("TestMultiSingle")

	for i := 0; i < b.N; i++ {
		Info(context.Background(), "测试打印日志", zap.String("name", "name"))
	}
}

func initContext() context.Context {
	ctx := context.Background()
	return ctx
}
