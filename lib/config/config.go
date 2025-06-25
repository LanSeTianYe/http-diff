package config

import (
	"time"
)

// Configs 配置信息
type Configs struct {
	App          App          `mapstructure:"app"`
	LoggerConfig LoggerConfig `mapstructure:"log"`
	FastHttp     FastHttp     `mapstructure:"fast_http"`
	DiffConfigs  []DiffConfig `mapstructure:"diff_configs"`
}

type App struct {
	Name string `mapstructure:"name"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	// Level 日志级别 zapcore.Level
	Level string `mapstructure:"level"`
	// Console 是否输出到控制台
	Console bool `mapstructure:"console"`
	// Path 日志文件路径
	Path string `mapstructure:"path"`
	// FileName 日志文件名
	FileName string `mapstructure:"file_name"`

	// MaxSize 日志文件最大大小，单位MB
	MaxSize int `mapstructure:"max_size"`
	// MaxAge 日志文件最大个数
	MaxBackups int `mapstructure:"max_backups"`
	// MaxAge 日志文件最大保存天数
	MaxAge int `mapstructure:"max_age"`
}

// FastHttp 配置
type FastHttp struct {
	// ReadTimeOut 从读缓冲区读取数据的超时时间，如果在调用的时候指定超时时间，则最短的一个会生效
	ReadTimeOut time.Duration `mapstructure:"read_time_out"`
	// WriteTimeOut 写入响应数据的超时时间，如果在调用的时候指定超时时间，则最短的一个会生效
	WriteTimeOut        time.Duration `mapstructure:"write_time_out"`
	MaxIdleConnDuration time.Duration `mapstructure:"max_idle_conn_duration"`
	MaxConnsPerHost     int           `mapstructure:"max_conns_per_host"`
	RetryTimes          int           `mapstructure:"retry_times"`
}

type DiffConfig struct {
	Name                 string        `mapstructure:"name"`
	Concurrency          int           `mapstructure:"concurrency"` // 并发数
	WaitTime             time.Duration `mapstructure:"wait_time"`   // 等待时间，每个请求完成之后等待的时间，可以用来限制请求的频率
	WorkDir              string        `mapstructure:"work_dir"`    // 工作目录
	Payload              string        `mapstructure:"payload"`     // 请求体内容,多个文件用逗号分割
	UrlA                 string        `mapstructure:"url_a"`
	UrlB                 string        `mapstructure:"url_b"`
	Method               string        `mapstructure:"method"`
	ContentType          string        `mapstructure:"content_type"`
	IgnoreFields         string        `mapstructure:"ignore_fields"`            // 忽略的字段，多个字段用逗号分割
	OutputShowNoDiffLine bool          `mapstructure:"output_show_no_diff_line"` // 输出是否展示没有差异的行，true 展示，false 不展示
	LogStatistics        bool          `mapstructure:"log_statistics"`           // 是否记录统计日志
	SuccessConditions    string        `mapstructure:"success_conditions"`       // 成功条件，多个条件用逗号分割
}
