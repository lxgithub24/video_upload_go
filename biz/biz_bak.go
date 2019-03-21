package biz

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"golivephoto/config"
	"golivephoto/grpcService/rpcClient"
	"golivephoto/models"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
)

func GetDLJ(stub, pic, mov, apptaskid, source string) (string, string, error) {
	gcids := rpcClient.Get_livephoto_gcid(stub)
	if !((pic == gcids[0] || pic == gcids[1]) && (mov == gcids[0] || mov == gcids[1])) {
		return "", "", errors.New("failed")
	}
	rearangeWidthHeight(mov)
	ok, insertId := models.LivePhotoCommit(pic, mov)
	if ok {
		fmt.Println(fileUrl)
		paramsMap := map[string]string{"url": fileUrl}
		paramsByte, _ := json.Marshal(paramsMap)
		paramsString := string(paramsByte)
		currentTime := strconv.FormatInt(time.Now().Unix(), 10)
		rand.Seed(time.Now().Unix())
		nonce := strconv.Itoa(rand.Intn(100000000))
		sign := formatSign(url, "POST", paramsString, payload)
		fmt.Println(url)
		res, _ := http.Post(url, "application/json", strings.NewReader(paramsString))
		fmt.Println(res.StatusCode)
		if res.StatusCode == 200 {
			resBody, _ := ioutil.ReadAll(res.Body)
			resJson, _ := simplejson.NewJson(resBody)
			dljString, _ := resJson.Get("dlj").String()
			return dljString, fileUrl, nil
		}
	}
	println("db op failed")
	return "", "", errors.New("failed")
}

func rearangeWidthHeight(gcid string) {
	qiniuUrl := config.QiniuHost + gcid + "?avinfo"
	res, err := http.Get(qiniuUrl)
	if err != nil {
		panic(err)
	}
	var width, height int64
	width, height = 480, 480
	var rotate bool = false
	if res.StatusCode == 200 {
		type Stream struct {
			Width  int64             `json:"width"`
			Height int64             `json:"height"`
			Tags   map[string]string `json:"tags"`
		}
		var streamsParam struct {
			Streams []*Stream `json:"streams"`
		}
		resBody, _ := ioutil.ReadAll(res.Body)
		err := json.Unmarshal([]byte(resBody), &streamsParam)
		if err != nil {
			fmt.Println(err)
			println("streams is wrong")
		}
		var tmpwidth, tmpheight int64
		tmprotate := "0"
		streamspaire := streamsParam.Streams
		for i := 0; i < len(streamspaire); i++ {
			fmt.Println(streamspaire[i])
			tmpwidth = streamspaire[i].Width
			tmpheight = streamspaire[i].Height
			tmprotate = streamspaire[i].Tags["rotate"]
		}
		tmprotateint, _ := strconv.Atoi(tmprotate)
		fmt.Println(tmprotateint)
		if tmprotateint >= 90 && tmprotateint <= 180 {
			rotate = true
		}
		if tmpwidth != 0 {
			width = tmpwidth
		}
		if tmpheight != 0 {
			height = tmpheight
		}
	}
	fmt.Println(rotate)
	if rotate {
		var tmp int64
		tmp = width
		width = height
		height = tmp
	}
	storagePlace := storage.EncodedEntry(config.QiniuBucket, gcid+"_mp4")
	// fops := fmt.Sprintf("avthumb/mp4/s/%dx%d/vb/500k|saveas/%s", width, height, storagePlace)
	fmt.Println(fops, "##########")
	mac := qbox.NewMac(config.QiniuAccessKey, config.QiniuSecretKey)
	cfg := storage.Config{
		UseHTTPS: true,
	}
	operationManager := storage.NewOperationManager(mac, &cfg)
	persistedId, err := operationManager.Pfop(config.QiniuBucket, gcid, fops, config.QiniuPipeline, config.QiniuCallback, true)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(persistedId)
}

func formatSign(uri, method, params string, payload map[string]string) string {
	baseString := fmt.Sprintf("%s&%s&", method, url.QueryEscape(uri))
	if len(payload) != 0 {
		var keys []string
		for k := range payload {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var sortedKeys []string
		for _, v := range keys {
			tmpString := fmt.Sprintf("%s%s%s", v, "%3D", payload[v])
			sortedKeys = append(sortedKeys, tmpString)
		}
		baseParam := strings.Join(sortedKeys, "%26")
		baseString += baseParam
	}
	if len(params) != 0 {
		baseString += ("%26" + params)
	}
	hmacHash := hmac.New(sha1.New, []byte("891bf2eae3b147deff8bc4f85e32d316"))
	hmacHash.Write([]byte(baseString))
	hmacSum := hmacHash.Sum(nil)
	sign := base64.URLEncoding.EncodeToString(hmacSum)
	return sign
}

type LivePhotoInfoRes struct {
	Cover_url string `json:"cover_url"`
	Play_url  string `json:"play_url"`
}

func LivePhotoInfo(id int64) (LivePhotoInfoRes, error) {
	var livephotoinfores = LivePhotoInfoRes{}
	res, err := models.LivePhotoInfo(id)
	if err != nil {
		return livephotoinfores, err
	}
	pic := formatUrl(res.Pic)
	mp4gcid := res.Mp4
	if res.Mp4 == "" {
		mp4gcid = res.Mov
	}
	mov := formatUrl(mp4gcid)
	livephotoinfores = LivePhotoInfoRes{Cover_url: pic, Play_url: mov}
	return livephotoinfores, nil
}

func formatUrl(gcid string) string {
	var CDN_EXPIRE int64 = 864000
	currenttime := strings.ToLower(strconv.FormatInt(time.Now().Unix()+CDN_EXPIRE, 16))
	source := fmt.Sprintf("%s/%s", QINIU_CDN_SECRET, gcid, currenttime)
	md5init := md5.New()
	md5init.Write([]byte(source))
	sign := strings.ToLower(hex.EncodeToString(md5init.Sum(nil)))
	url := fmt.Sprintf("%s%s?sign=%s&t=%s", QINIU_LIVEPHOTO_DOWNLOAD_URI, gcid, sign, currenttime)
	return url
}

func Callbackmp4(mp4, mov string) error {
	err := models.Callbackmp4(mp4, mov)
	return err
}
