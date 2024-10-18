package main

import (
	"fmt"
	"github.com/echo/hade/framework/gin"
)

func Test1() gin.HandlerFunc {
	println("middle test1")
	return func(c *gin.Context) {
		fmt.Println("middleware pre test1")
		c.Next() // 调用Next往下调用，会自增contxt.index
		fmt.Println("middleware post test1")
	}
}

func Test2() gin.HandlerFunc {
	println("middle test2")
	return func(c *gin.Context) {
		fmt.Println("middleware pre test2")
		c.Next() // 调用Next往下调用，会自增contxt.index
		fmt.Println("middleware post test2")
	}
}

func Test3() gin.HandlerFunc {
	println("middle test3")
	return func(c *gin.Context) {
		fmt.Println("middleware pre test3")
		c.Next() // 调用Next往下调用，会自增contxt.index
		fmt.Println("middleware post test3")
	}
}
