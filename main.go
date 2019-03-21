package main

import (
	"flag"
	"fmt"
	"golivephoto/api"
	"golivephoto/config"
	"golivephoto/models"
	"golivephoto/tools"
	"time"

	"github.com/fvbock/endless"
	raven "github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
)

var (
	addr = flag.String("listen-address", ":18001", "The address to listen on for HTTP requests.")
)

func main() {
	fmt.Println("###################")
	fmt.Println(time.Now())
	flag.Parse()
	config.InitConfig()
	models.InitModels()

	router := gin.New()
	client, err := raven.New(config.SentryDsn)
	if err != nil {
		println(err)
	}
	router.Use(tools.Logger())
	router.Use(tools.CatchPanicError(client))
	router.Use(tools.Prometheus())
	router.GET("/live_photo/api/metrics", tools.MetricsFunc)
	//http://127.0.0.1:18001/metrics
	router.POST("/live_photo/commit", api.Commit)
	//curl -XPOST "http://127.0.0.1:18001/live_photo/commit" -H "Content-Type:application/json" -d '{"stub":"123","pic":"8afac63ebaf62455c4ad01a5c3f3e10d4937a3f8","mov":"c4a446fa87c4e830813a6dada5781b98ecbab343","apptaskid":"46e0cb5bd8683e4b9e6d2944f4968235_3e54ba067ee43b1d84e86864eacfc4ac_1518442367193","source":"wechat"}'
	router.GET("/live_photo/info", api.Info)
	//http://127.0.0.1:18001/live_photo/info?id=1
	router.POST("/live_photo/api/callback", api.Callback)
	//curl -XPOST "http://127.0.0.1:18001/live_photo/api/callback" -H "Content-Type:application/json" -d '{"inputKey":"123","items":[{"key":""}]}'
	endless.ListenAndServe(*addr, router)
}
