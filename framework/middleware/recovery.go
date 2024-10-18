package middleware

import (
	"github.com/Superdanda/hade/framework/gin"
	"net/http"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.ISetStatus(http.StatusInternalServerError).IJson(err)
			}
		}()
		c.Next()
	}
}
