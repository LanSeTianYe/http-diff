package config

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Init 初始化配置，把配置文件解析到结构体中
func Init(configFile string, configStruct any) error {
	viper.SetConfigFile(configFile)

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("read config file error, config file:%s, err: %w", configFile, err)
	}

	viperHookFunc := mapstructure.ComposeDecodeHookFunc(
		// 字符串转时间间隔 1s 1m 1h 1d
		mapstructure.StringToTimeDurationHookFunc(),
		// 字符串转字符串数组 1,2,3 => [1,2,3]
		mapstructure.StringToSliceHookFunc(","),
	)

	err = viper.Unmarshal(configStruct, viper.DecodeHook(viperHookFunc))
	if err != nil {
		return fmt.Errorf("unmarshal config file error, config file:%s, config data:%v, err: %w", configFile, configStruct, err)
	}

	fmt.Printf("config info, config path:%#v, config:%#v\n", configFile, configStruct)

	return nil
}
