package kernel

import (
	"github.com/Superdanda/hade/framework/gin"
	"net/http"
)

type HadeKernelService struct {
	engine *gin.Engine
}

func NewHadeKernelService(params ...interface{}) (interface{}, error) {
	httpEngine := params[0].(*gin.Engine)
	return &HadeKernelService{engine: httpEngine}, nil
}

func (kernelService *HadeKernelService) HttpEngine() http.Handler {
	return kernelService.engine
}
