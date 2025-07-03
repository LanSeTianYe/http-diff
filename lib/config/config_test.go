package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	var conf = &Configs{}
	err := Init("./data/config.toml", conf)
	assert.Nil(t, err)

	assert.NotNil(t, conf)
	assert.NotNil(t, conf.App)
	assert.NotNil(t, conf.LoggerConfig)

	assert.Equal(t, "test", conf.App.Name)

	assert.Equal(t, true, conf.LoggerConfig.Console)
	assert.Equal(t, "DEBUG", conf.LoggerConfig.Level)
	assert.Equal(t, "./", conf.LoggerConfig.Path)
	assert.Equal(t, "server.log", conf.LoggerConfig.FileName)
	assert.Equal(t, 100, conf.LoggerConfig.MaxSize)
	assert.Equal(t, 30, conf.LoggerConfig.MaxBackups)
	assert.Equal(t, 15, conf.LoggerConfig.MaxAge)

	assert.Equal(t, time.Millisecond*500, conf.FastHttp.ReadTimeOut)
	assert.Equal(t, time.Millisecond*500, conf.FastHttp.WriteTimeOut)
	assert.Equal(t, time.Hour, conf.FastHttp.MaxIdleConnDuration)
	assert.Equal(t, 512, conf.FastHttp.MaxConnsPerHost)
	assert.Equal(t, 2, conf.FastHttp.RetryTimes)

	assert.Equal(t, "task_1", conf.DiffConfigs[0].Name)
	assert.Equal(t, 10, conf.DiffConfigs[0].Concurrency)
	assert.Equal(t, time.Duration(0), conf.DiffConfigs[0].WaitTime)
	assert.Equal(t, "./data", conf.DiffConfigs[0].WorkDir)
	assert.Equal(t, "payload_task_1.txt", conf.DiffConfigs[0].Payload)
	assert.Equal(t, "https://example.com/url_a", conf.DiffConfigs[0].UrlA)
	assert.Equal(t, "https://example.com/url_b", conf.DiffConfigs[0].UrlB)
	assert.Equal(t, "GET", conf.DiffConfigs[0].Method)
	assert.Equal(t, "application/json", conf.DiffConfigs[0].ContentType)
	assert.Equal(t, "field_a", conf.DiffConfigs[0].IgnoreFields)
	assert.True(t, conf.DiffConfigs[0].OutputShowNoDiffLine)
	assert.False(t, conf.DiffConfigs[0].LogStatistics)
	assert.Equal(t, "stat=1,code=0", conf.DiffConfigs[0].SuccessConditions)

	assert.Equal(t, "task_2", conf.DiffConfigs[1].Name)
	assert.Equal(t, 5, conf.DiffConfigs[1].Concurrency)
	assert.Equal(t, time.Millisecond*1000, conf.DiffConfigs[1].WaitTime)
	assert.Equal(t, "./data", conf.DiffConfigs[1].WorkDir)
	assert.Equal(t, "payload_task_2.txt", conf.DiffConfigs[1].Payload)
	assert.Equal(t, "https://example.com/url_a", conf.DiffConfigs[1].UrlA)
	assert.Equal(t, "https://example.com/url_b", conf.DiffConfigs[1].UrlB)
	assert.Equal(t, "POST", conf.DiffConfigs[1].Method)
	assert.Equal(t, "application/json", conf.DiffConfigs[1].ContentType)
	assert.Equal(t, "field_b", conf.DiffConfigs[1].IgnoreFields)
	assert.False(t, conf.DiffConfigs[1].OutputShowNoDiffLine)
	assert.True(t, conf.DiffConfigs[1].LogStatistics)
	assert.Equal(t, "stat=1,code=1", conf.DiffConfigs[1].SuccessConditions)
}
