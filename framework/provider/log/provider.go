package log

import (
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
	"github.com/Superdanda/hade/framework/provider/log/formatter"
	"github.com/Superdanda/hade/framework/provider/log/services"
	"io"
	"strings"
)

type HadeLogServiceProvider struct {
	Driver string // Driver

	Level contract.LogLevel // 日志级别

	Formatter contract.Formatter // 日志输出格式方法

	CtxFielder contract.CtxFielder // 日志context上下文信息获取函数

	Output io.Writer // 日志输出信息
}

func (l HadeLogServiceProvider) Register(container framework.Container) framework.NewInstance {
	if l.Driver == "" {
		tcs, err := container.Make(contract.ConfigKey)
		if err != nil {
			// 默认使用console
			return services.NewHadeConsoleLog
		}

		cs := tcs.(contract.Config)
		l.Driver = strings.ToLower(cs.GetString("log.driver"))
	}

	// 根据driver的配置项确定
	switch l.Driver {
	case "single":
		return services.NewHadeSingleLog
	case "rotate":
		return services.NewHadeRotateLog
	case "console":
		return services.NewHadeConsoleLog
	case "custom":
		return services.NewHadeCustomLog
	default:
		return services.NewHadeConsoleLog
	}
}

func (l HadeLogServiceProvider) Boot(container framework.Container) error {
	return nil
}

func (l HadeLogServiceProvider) IsDefer() bool {
	return false
}

// Params 定义要传递给实例化方法的参数
func (l HadeLogServiceProvider) Params(container framework.Container) []interface{} {
	// 获取configService
	configService := container.MustMake(contract.ConfigKey).(contract.Config)

	// 设置参数formatter
	if l.Formatter == nil {
		l.Formatter = formatter.TextFormatter
		if configService.IsExist("log.formatter") {
			v := configService.GetString("log.formatter")
			if v == "json" {
				l.Formatter = formatter.JsonFormatter
			} else if v == "text" {
				l.Formatter = formatter.TextFormatter
			}
		}
	}

	if l.Level == contract.UnknownLevel {
		l.Level = contract.InfoLevel
		if configService.IsExist("log.level") {
			l.Level = logLevel(configService.GetString("log.level"))
		}
	}

	// 定义5个参数
	return []interface{}{container, l.Level, l.CtxFielder, l.Formatter, l.Output}
}

func (l HadeLogServiceProvider) Name() string {
	return contract.LogKey
}

func logLevel(config string) contract.LogLevel {
	switch strings.ToLower(config) {
	case "panic":
		return contract.PanicLevel
	case "fatal":
		return contract.FatalLevel
	case "error":
		return contract.ErrorLevel
	case "warn":
		return contract.WarnLevel
	case "info":
		return contract.InfoLevel
	case "debug":
		return contract.DebugLevel
	case "trace":
		return contract.TraceLevel
	}
	return contract.UnknownLevel
}
