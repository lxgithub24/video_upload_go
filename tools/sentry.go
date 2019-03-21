package tools

import (
	"errors"
	"fmt"
	"net/http"

	raven "github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
)

func CatchPanicError(client *raven.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				errStr := fmt.Sprint(err)
				endPoint := c.Request.URL.Path
				flags := map[string]string{"endPoint": endPoint}
				packet := raven.NewPacket(errStr, raven.NewException(errors.New(errStr), raven.NewStacktrace(3, 3, nil)))
				client.Capture(packet, flags)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
