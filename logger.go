package gow

import (
	"fmt"
	"time"
)

// [gow] 2020/07/01 - 14:55:52 | 200 |      44.961µs |       127.0.0.1 | GET      "/article/1"
// Logger
func Logger() HandlerFunc {
	return func(c *Context) {
		t := time.Now()
		c.Next()
		fmt.Printf("[%s] %s | %-3d | %-15s| %-5s | %-10s | %s \n", c.engine.AppName, time.Now().Format("2006/01/02 15:04:05"), c.StatusCode, c.IP(), c.Method, time.Since(t), c.Path)
	}
}
