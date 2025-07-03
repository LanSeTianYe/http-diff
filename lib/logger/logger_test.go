package logger

import (
	"context"
	"testing"

	"http-diff/lib/config"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func initLogger(t assert.TestingT, name string) {
	configs := config.Configs{}
	err := config.Init("./data/config.toml", &configs)
	assert.Nil(t, err)

	Init(name, configs.LoggerConfig)
}

func TestW(t *testing.T) {
	initLogger(t, "TestF")

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
	initLogger(t, "TestMultiSingle")

	times := 1024
	for i := 0; i < times; i++ {
		Info(context.Background(), "测试打印日志", zap.String("name", "name"))
	}
}

// BenchmarkLogger-8   	  374352	      3036 ns/op
func BenchmarkLogger(b *testing.B) {
	initLogger(b, "TestMultiSingle")

	for i := 0; i < b.N; i++ {
		Info(context.Background(), "测试打印日志", zap.String("name", "name"))
	}
}

func initContext() context.Context {
	ctx := context.Background()
	return ctx
}
