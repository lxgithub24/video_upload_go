package tools

import (
	"fmt"
	"time"

	log "github.com/cihub/seelog"
	"github.com/gin-gonic/gin"
)

var ()

func Logger() gin.HandlerFunc {
	L4g, Err := log.LoggerFromConfigAsFile("/data/apps/golivephoto/golivephoto/logging.xml")
	// println("#######", L4g, Err)
	defer L4g.Flush()
	if Err != nil {
		print(Err)
	}
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()
		end := time.Now()
		latency := end.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		comment := c.Errors.ByType(gin.ErrorTypePrivate).String()
		lginfo := fmt.Sprintf("%3d | %13v | %s | %-15s | %-7s %s %s", statusCode, latency, start.Format("2006-01-02 15:04:05"), clientIP, method, path, comment)
		L4g.Infof("%s", lginfo)
	}
}
