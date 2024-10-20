package distributed

import (
	"errors"
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
	"github.com/gofrs/flock"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type LocalDistributedService struct {
	container framework.Container //服务容器
}

func NewLocalDistributedService(params ...interface{}) (interface{}, error) {
	if len(params) != 1 {
		return nil, errors.New("param error")
	}
	// 有两个参数，一个是容器，一个是baseFolder
	container := params[0].(framework.Container)
	return &LocalDistributedService{container: container}, nil
}

// Select 为分布式选择器
func (s LocalDistributedService) Select(serviceName string, appID string, holdTime time.Duration) (selectAppID string, err error) {
	appService := s.container.MustMake(contract.AppKey).(contract.App)
	runtimeFolder := appService.RuntimeFolder()
	lockFile := filepath.Join(runtimeFolder, "distribute_"+serviceName)

	// 使用 flock 创建文件锁
	fileLock := flock.New(lockFile)

	// 尝试加独占锁
	locked, err := fileLock.TryLock()
	if err != nil {
		return "", err
	}

	if !locked {
		// 如果未能获取锁，读取文件中的 appID
		selectAppIDByt, err := ioutil.ReadFile(lockFile)
		if err != nil {
			return "", err
		}
		return string(selectAppIDByt), nil
	}

	// 在选举有效时间内其他节点不能再抢占
	go func() {
		defer func() {
			// 释放文件锁
			_ = fileLock.Unlock()
			// 删除锁文件
			_ = os.Remove(lockFile)
		}()

		// 创建一个计时器以保持选举的有效性
		timer := time.NewTimer(holdTime)
		<-timer.C
	}()

	// 将抢占到的 appID 写入文件
	if err := os.WriteFile(lockFile, []byte(appID), 0666); err != nil {
		return "", err
	}

	return appID, nil
}
