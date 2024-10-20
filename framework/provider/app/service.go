package app

import (
	"fmt"
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/util"
	"github.com/google/uuid"
	"path/filepath"
)

type HadeApp struct {
	container  framework.Container
	baseFolder string
	appId      string

	configMap map[string]string // 配置加载
}

func NewHadeApp(params ...interface{}) (interface{}, error) {
	if len(params) != 2 {
		return nil, fmt.Errorf("HadeApp NewHadeApp expects 2 parameters")
	}
	// 有两个参数，一个是容器，一个是baseFolder
	container := params[0].(framework.Container)
	baseFolder := params[1].(string)
	appId := uuid.New().String()
	configMap := make(map[string]string)
	return &HadeApp{container: container, baseFolder: baseFolder, appId: appId, configMap: configMap}, nil
}

func (h HadeApp) BaseFolder() string {
	if h.baseFolder != "" {
		return h.baseFolder
	}
	//var baseFolder string
	//flag.StringVar(&baseFolder, "base_folder", "", "base_folder 参数, 默认为当前路径")
	//flag.Parse()
	//if baseFolder != "" {
	//	return baseFolder
	//}
	return util.GetExecDirectory()
}

func (h HadeApp) Version() string {
	return "0.0.1"
}

func (h HadeApp) StorageFolder() string {
	return filepath.Join(h.BaseFolder(), "storage")
}

func (h HadeApp) ConfigFolder() string {
	return filepath.Join(h.BaseFolder(), "config")
}

func (h HadeApp) LogFolder() string {
	if val, ok := h.configMap["log_folder"]; ok {
		return val
	}
	return filepath.Join(h.StorageFolder(), "log")
}

func (h HadeApp) ProviderFolder() string {
	return filepath.Join(h.BaseFolder(), "provider")
}

func (h HadeApp) MiddlewareFolder() string {
	return filepath.Join(h.HttpFolder(), "middleware")
}

func (h HadeApp) CommandFolder() string {
	return filepath.Join(h.ConsoleFolder(), "command")
}

func (h HadeApp) RuntimeFolder() string {
	return filepath.Join(h.StorageFolder(), "runtime")
}

func (h HadeApp) TestFolder() string {
	return filepath.Join(h.BaseFolder(), "test")
}

func (h HadeApp) HttpFolder() string {
	return filepath.Join(h.BaseFolder(), "http")
}

func (h HadeApp) ConsoleFolder() string {
	return filepath.Join(h.BaseFolder(), "console")
}

func (h HadeApp) AppId() string {
	return h.appId
}

func (h HadeApp) LoadAppConfig(mapString map[string]string) {
	for key, val := range mapString {
		h.configMap[key] = val
	}
}
