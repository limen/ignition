package middlewares

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

func AccessLogHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fmt.Printf(
			"[[access]] %s [IP=%s] [method=%s] [URI=%s]\n",
			formatTime(time.Now()),
			ctx.ClientIP(),
			ctx.Request.Method,
			ctx.Request.RequestURI,
		)
		ctx.Next()
	}
}

func formatTime(t time.Time) string {
	return t.Format("2006/01/02 - 15:04:05")
}
