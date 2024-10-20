package demo

import (
	"github.com/Superdanda/hade/framework/contract"
	"github.com/Superdanda/hade/framework/gin"
)

type DemoApi struct {
}

func Register(r *gin.Engine) error {
	r.GET("/demo/demo", Demo)
	return nil
}

func Demo(c *gin.Context) {
	configService := c.MustMake(contract.ConfigKey).(contract.Config)
	log := c.MustMake(contract.LogKey).(contract.Log)
	password := configService.GetString("database.mysql.password")
	log.Info(c, "ceshiceshi", map[string]interface{}{})
	c.JSON(200, password)
}
