package contract

import "net/http"

const KernelKey = "hade:kernel"

type Kernel interface {
	// HttpEngine http.Handler结构，作为net/http框架使用, 实际上是gin.Engine
	HttpEngine() http.Handler
}
