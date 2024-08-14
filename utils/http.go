package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	Log "github.com/wellmoon/go/logger"
	"github.com/wellmoon/go/zjson"
)

type Downloader struct {
	io.Reader
	Total   int64
	Current int64
}

func (d *Downloader) Read(p []byte) (n int, err error) {
	n, err = d.Reader.Read(p)
	d.Current += int64(n)
	var unit string = "M"
	var downloaded = d.Total / 1024 / 1024
	if downloaded <= 0 {
		downloaded = d.Total / 1024
		unit = "K"
		if downloaded <= 0 {
			downloaded = d.Total
			unit = "B"
		}
	}
	if downloaded < 0 {
		downloaded = 0
	}
	if d.Current >= d.Total {
		fmt.Printf("\r下载完成，共[%v]%v                \n", downloaded, unit)
	} else {
		fmt.Printf("\r正在下载，共[%v]%v，进度：%.2f%%", downloaded, unit, float64(d.Current*10000/d.Total)/100)
	}
	return
}

func DownloadFile(url string, locTarget string) (string, error) {

	_, fileName := filepath.Split(url)
	resp, err := http.Get(url)
	if err != nil {
		Log.Error("http get {} error: {}", err, url, err)
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	isDir := IsDir(locTarget)
	if isDir {
		locTarget = locTarget + fileName
	}
	if !PathExists(locTarget) {
		Log.Debug("download file [{}] to local...", fileName)
	} else {
		locFile, err := os.OpenFile(locTarget, os.O_RDONLY, os.ModePerm)
		if err != nil {
			return "", err
		}
		locFileStat, _ := locFile.Stat()
		if locFileStat.Size() != resp.ContentLength {
			Log.Trace("本地文件已存在，但与即将下载的文件大小不一致，重新下载。[{}]", locTarget)
		} else {
			Log.Trace("本地文件已存在且文件大小一致，无需下载。如需重新下载，请先删除本地文件[{}]", locTarget)
			return "exist", nil
		}
	}
	Log.Debug("download file, total [{}M]：{}", resp.ContentLength/1024/1024, locTarget)
	file, _ := os.Create(locTarget)
	defer func() {
		_ = file.Close()
	}()
	downloader := &Downloader{
		Reader: resp.Body,
		Total:  resp.ContentLength,
	}
	if _, err := io.Copy(file, downloader); err != nil {
		Log.Fatal("io.Copy error, {}", err)
		return "", err
	}
	return fileName, nil
}

func IsDir(path string) bool {
	s, err := os.Stat(path)

	if err != nil {
		return false
	}
	return s.IsDir()

}

func SendReqWithTimeout(url string, requestType string, params map[string]interface{}, headers map[string]string, timeout int) (string, map[string]string, map[string]string, error) {
	return SendReqWithProxy(url, requestType, params, headers, nil, timeout)
}

func SendReq(url string, requestType string, params map[string]interface{}, headers map[string]string) (string, map[string]string, map[string]string, error) {
	return SendReqWithProxy(url, requestType, params, headers, nil, 15)
}

func SendReqRaw(url string, params map[string]interface{}) (string, error) {
	return SendReqRawWithHeader(url, params, nil)
}
func SendReqRawWithHeader(url string, params map[string]interface{}, headers map[string]string) (string, error) {

	b1, _ := json.Marshal(&params)
	return SendBytes(url, b1, headers)
}

func SendBytes(url string, inputBytes []byte, headers map[string]string) (string, error) {
	client := http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(inputBytes))
	if err != nil {
		log.Println("err")
		return "", err
	}
	if len(headers) > 0 {
		for key, val := range headers {
			req.Header.Add(key, val)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("err")
		return "", err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("err")
		return "", err
	}
	return string(b), nil
}

func SendReqRawReturnBytes(url string, params map[string]string, headers map[string]interface{}) ([]byte, error) {
	client := http.Client{}

	b1, _ := json.Marshal(&params)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b1))
	if err != nil {
		log.Println("err")
		return nil, err
	}
	if len(headers) > 0 {
		for key, val := range headers {
			req.Header.Add(key, zjson.ToStr(val))
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("err")
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("err")
		return nil, err
	}
	return b, nil
}

func SendReqWithProxy(url string, requestType string, params map[string]interface{}, headers map[string]string, proxy *url.URL, timeoutSeconds int) (string, map[string]string, map[string]string, error) {
	// client := &http.Client{
	// 	Timeout: time.Duration(timeoutSeconds) * time.Second,
	// }
	if timeoutSeconds <= 0 {
		timeoutSeconds = 10
	}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()
	if proxy != nil {
		http.DefaultClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
	}
	reqType := "GET"
	requestType = strings.ToUpper(requestType)
	if requestType == "POST" {
		reqType = requestType
	}

	var paramStr string
	if len(params) > 0 {
		var r http.Request
		r.ParseForm()
		for key, val := range params {
			r.Form.Add(key, zjson.ToStr(val))
		}
		paramStr = strings.TrimSpace(r.Form.Encode())
	}

	request, err := http.NewRequestWithContext(ctx, reqType, url, strings.NewReader(paramStr))
	if err != nil {
		return "", nil, nil, err
	}

	if len(headers) > 0 {
		for key, val := range headers {
			request.Header.Add(key, val)
		}
	}
	if requestType == "POST" {
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	//处理返回结果
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", nil, nil, err
	}
	defer response.Body.Close()
	respHeaders := response.Header
	respHeaderMap := make(map[string]string)
	for k := range respHeaders {
		respHeaderMap[k] = respHeaders.Get(k)
	}

	cookies := response.Cookies()
	respCookiesMap := make(map[string]string)
	for _, c := range cookies {
		respCookiesMap[c.Name] = c.Value
	}

	resBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", nil, nil, err
	}
	return string(resBytes), respHeaderMap, respCookiesMap, nil

}

func ResumeShortUrl(url string) string {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	reqType := "GET"

	request, err := http.NewRequest(reqType, url, nil)
	if err != nil {
		return url
	}

	//处理返回结果
	response, err := client.Do(request)
	if err != nil {
		return url
	}
	defer response.Body.Close()
	return response.Request.URL.String()
}
