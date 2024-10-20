package contract

import "time"

const DistributedKey = "hade:distributed"

type Distributed interface {
	Select(serviceName string, appId string, holdTime time.Duration) (selectAppId string, err error)
}
