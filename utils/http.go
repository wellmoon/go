package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	Log "github.com/wellmoon/go/logger"
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
		Log.Trace("http get", err)
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

func SendReq(url string, requestType string, params map[string]string, headers map[string]string) (string, map[string]string, map[string]string, error) {
	client := &http.Client{}
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
			r.Form.Add(key, val)
		}
		paramStr = strings.TrimSpace(r.Form.Encode())
	}

	request, err := http.NewRequest(reqType, url, strings.NewReader(paramStr))
	if err != nil {
		Log.Error("NewRequest error : {}", err)
		return "", nil, nil, err
	}

	if len(headers) > 0 {
		for key, val := range headers {
			request.Header.Add(key, val)
		}
	}

	//处理返回结果
	response, err := client.Do(request)
	if err != nil {
		Log.Error("client Do error : {}", err)
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
		Log.Error("read resp error : {}", err)
		return "", nil, nil, err
	}
	return string(resBytes), respHeaderMap, respCookiesMap, nil

}
