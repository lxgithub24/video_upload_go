package api

import (
	"fmt"
	"golivephoto/biz"
	"golivephoto/config"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CommitParams struct {
	Stub      string `json:"stub"`
	Pic       string `json:"pic"`
	Mov       string `json:"mov"`
	Apptaskid string `json:"apptaskid"`
	Source    string `json:"source"`
}

func Commit(c *gin.Context) {
	var commitParam CommitParams
	c.BindJSON(&commitParam)
	stub := commitParam.Stub
	pic := commitParam.Pic
	mov := commitParam.Mov
	apptaskid := commitParam.Apptaskid
	source := commitParam.Source
	fmt.Println(commitParam)
	if stub == "" || pic == "" || mov == "" {
		c.JSON(http.StatusBadRequest, gin.H{"result": config.CODE_BAD_PARAMS})
		return
	}

	short_url, origin_url, err := biz.GetDLJ(stub, pic, mov, apptaskid, source)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": config.CODE_BAD_PARAMS})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"result":     "ok",
		"short_url":  short_url,
		"origin_url": origin_url,
	})
}

func Info(c *gin.Context) {
	IdString := c.Query("id")
	fmt.Println(IdString)
	IdInt, _ := strconv.Atoi(IdString)
	fmt.Println(IdInt)
	Id := int64(IdInt)
	fmt.Println(Id)
	if Id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"result": "bad params"})
		return
	}
	infoJson, err := biz.LivePhotoInfo(Id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": "bad params"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"result":     "ok",
		"live_photo": infoJson,
	})
}

type CallbackParams struct {
	Code     int            `json:"code"`
	Items    []*ItemsParams `json:"items"`
	InputKey string         `json:"inputKey"`
}

type ItemsParams struct {
	Key string `json:"key"`
}

func Callback(c *gin.Context) {
	var callbackParam CallbackParams

	if err := c.BindJSON(&callbackParam); err != nil {
		fmt.Println("Callback BindJSON err: ", err)
	}

	codeParam := callbackParam.Code
	mp4 := callbackParam.Items[0].Key
	mov := callbackParam.InputKey
	if codeParam != 0 || mov == "" || mp4 == "" {
		c.JSON(http.StatusBadRequest, gin.H{"result": "bad params"})
		return
	}
	err := biz.Callbackmp4(mp4, mov)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": "bad params"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"result": "ok",
	})
}
