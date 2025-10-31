package gintool

import (
	"time"

	"github.com/google/uuid"
	"github.com/hellobchain/wswlog/wlogging"

	"github.com/gin-gonic/gin"
)

var logClient = wlogging.MustGetFileLoggerWithoutName(nil)

// 请求日志汇总信息
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := uuid.New().ID()
		// 开始时间
		start := time.Now()
		// path
		path := c.Request.URL.Path
		// ip
		clientIP := c.ClientIP()
		// 方法
		method := c.Request.Method
		// 处理请求
		c.Next()
		// 结束时间
		end := time.Now()
		// 执行时间
		latency := end.Sub(start)
		// 状态
		statusCode := c.Writer.Status()
		logClient.Infof("| %10d | %3d | %13v | %15s | %s  %s |",
			requestId,
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)
	}
}
