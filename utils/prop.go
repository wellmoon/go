package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	Log "github.com/wellmoon/go/logger"
)

type PathVo struct {
	MainPath    string
	ConfPath    string
	DockerPath  string
	SourcesPath string
	DeployPath  string
	FlowPath    string
	FrontPath   string
	AcsPath     string
	Sep         string
}

var path *PathVo

func CurPath() string {
	curPath, _ := os.Getwd()
	executable, _ := os.Executable()
	dir := filepath.Dir(executable)
	var mainPath string
	if strings.HasPrefix(dir, curPath) {
		mainPath = dir
	} else {
		mainPath = curPath
	}
	return mainPath
}

func Path() *PathVo {
	if path != nil {
		return path
	}
	mainPath := CurPath()

	sep := string(os.PathSeparator)

	// 创建必要文件夹
	mainPath = mainPath + sep
	sourcesPath := mainPath + "sources" + sep
	confExist := PathExists(sourcesPath)
	if !confExist {
		os.Mkdir(mainPath+"conf", os.ModePerm)
		os.Mkdir(mainPath+"docker", os.ModePerm)
		os.Mkdir(sourcesPath, os.ModePerm)
		os.Mkdir(sourcesPath+"deploy", os.ModePerm)
		os.Mkdir(sourcesPath+"flow", os.ModePerm)
		os.Mkdir(sourcesPath+"front", os.ModePerm)
		os.Mkdir(sourcesPath+"acs", os.ModePerm)
	}
	pathVo := &PathVo{
		MainPath:    mainPath,
		ConfPath:    mainPath + "conf" + sep,
		DockerPath:  mainPath + "docker" + sep,
		SourcesPath: sourcesPath,
		DeployPath:  sourcesPath + "deploy" + sep,
		FlowPath:    sourcesPath + "flow" + sep,
		FrontPath:   sourcesPath + "front" + sep,
		AcsPath:     sourcesPath + "acs" + sep,
		Sep:         sep,
	}
	path = pathVo
	return path
}

func InitConfig() map[string]string {

	// 创建leridge.conf文件
	confFile := Path().ConfPath + "leridge.conf"
	confFileExist := PathExists(confFile)
	if !confFileExist {
		createConf(Path().ConfPath)
	}

	config := GetConf(confFile)
	Log.Debug("当前程序路径:%v", Path().MainPath)
	Log.Debug("服务器信息如下，如有变更，修改[conf]目录下的[leridge.conf]文件:")
	Log.Debug("IP:%v\n", config["ip"])
	Log.Debug("PORT:%v\n", config["port"])
	Log.Debug("飞流用户名:%v\n", config["flowuser"])
	Log.Debug("飞流用户密码:%v\n", config["flowpass"])
	Log.Debug("------------------------------------")
	return config
}

func GetConf(confFile string) map[string]string {
	config := make(map[string]string)
	f, err := os.Open(confFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		b, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		s := strings.TrimSpace(string(b))
		index := strings.Index(s, "=")
		if index < 0 {
			continue
		}
		key := strings.TrimSpace(s[:index])
		if len(key) == 0 {
			continue
		}
		value := strings.TrimSpace(s[index+1:])
		if len(value) == 0 {
			continue
		}
		config[key] = value
	}

	return config
}

func createConf(confPath string) {
	var ip, port, password string
	fmt.Println("没找到服务器配置，请输入服务器IP地址：")
	fmt.Scanln(&ip)
	fmt.Println("请输入端口号（不输入默认为22）：")
	fmt.Scanln(&port)
	fmt.Println("请输入root的密码：")
	fmt.Scanln(&password)

	if len(port) == 0 {
		port = "22"
	}
	sep := string(os.PathSeparator)

	// 创建root相关配置
	f, err := os.Create(confPath + sep + "leridge.conf") //创建文件
	if err != nil {
		fmt.Println("创建配置文件失败", err)
	}
	defer f.Close()
	var _, err1 = io.WriteString(f, "ip="+ip+"\n")
	check(err1)
	_, err1 = io.WriteString(f, "port="+port+"\n")
	check(err1)
	_, err1 = io.WriteString(f, "pass="+password+"\n")
	check(err1)
	_, err1 = io.WriteString(f, "flowuser=flow\n")
	check(err1)
	_, err1 = io.WriteString(f, "flowpass=leridge123\n")
	check(err1)
}

func check(e error) {
	if e != nil {
		Log.Error("error, %v", e)
	}
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
