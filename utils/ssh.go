package utils

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"time"

	Log "github.com/wellmoon/go/logger"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Cli struct {
	IP         string      //IP地址
	Username   string      //用户名
	Password   string      //密码
	Port       int         //端口号
	client     *ssh.Client //ssh客户端
	LastResult string      //最近一次Run的结果
}

//创建命令行对象
//@param ip IP地址
//@param username 用户名
//@param password 密码
//@param port 端口号,默认22
// func New(ip string, username string, password string, port ...int) *Cli {
// 	cli := new(Cli)
// 	cli.IP = ip
// 	cli.Username = username
// 	cli.Password = password
// 	if len(port) <= 0 {
// 		cli.Port = 22
// 	} else {
// 		cli.Port = port[0]
// 	}
// 	return cli
// }

func New(conf map[string]string) *Cli {
	cli := new(Cli)
	cli.IP = conf["ip"]
	cli.Username = "root"
	cli.Password = conf["pass"]
	port := conf["port"]
	if len(port) <= 0 {
		cli.Port = 22
	} else {
		cli.Port, _ = strconv.Atoi(port)
	}
	return cli
}

func NewFLowCli(conf map[string]string) *Cli {
	cli := new(Cli)
	cli.IP = conf["ip"]
	cli.Username = conf["flowuser"]
	cli.Password = conf["flowpass"]
	port := conf["port"]
	if len(port) <= 0 {
		cli.Port = 22
	} else {
		cli.Port, _ = strconv.Atoi(port)
	}
	return cli
}

func NewDeployServer(ip string, port int, username string, pass string) *Cli {
	cli := new(Cli)
	cli.IP = ip
	cli.Username = username
	cli.Password = pass
	cli.Port = port
	return cli
}

func NewOther(ip string, port int, username string, pass string) *Cli {
	cli := new(Cli)
	cli.IP = ip
	cli.Username = username
	cli.Password = pass
	cli.Port = port
	return cli
}

func checkClient(c *Cli) error {
	if c.client == nil {
		if err := c.connect(); err != nil {
			Log.Error("connect remote server err, %v", err)
			return err
		}
	}
	return nil
}

//执行shell
//@param shell shell脚本命令
func (c Cli) Run(shell string) string {
	err := checkClient(&c)
	if err != nil {
		return err.Error()
	}
	session, err := c.client.NewSession()
	if err != nil {
		Log.Error("new session err, %v", err)
		return err.Error()
	}
	defer session.Close()
	buf, _ := session.CombinedOutput(shell)
	c.LastResult = string(buf)
	return c.LastResult
}

//执行shell
//@param shell shell脚本命令
func (c Cli) RunMulti(shells []string) string {
	err := checkClient(&c)
	if err != nil {
		return err.Error()
	}
	session, err := c.client.NewSession()
	if err != nil {
		Log.Error("new session err, %v", err)
		return err.Error()
	}
	stdinBuf, err := session.StdinPipe()
	if err != nil {
		Log.Error("StdinPipe err, %v", err)
		return err.Error()
	}
	var outbt, errbt bytes.Buffer
	session.Stdout = &outbt
	session.Stderr = &errbt
	err = session.Shell()
	if err != nil {
		Log.Error("session Shell err, %v", err)
		return err.Error()
	}
	shells = append(shells, "exit")
	for _, c := range shells {
		c = c + "\n"
		stdinBuf.Write([]byte(c))
	}
	session.Wait()
	c.LastResult = outbt.String()
	session.Close()
	return c.LastResult
}

type UpdateFile struct {
	File    *os.File
	Total   int64
	Current int64
	Finish  bool
}

type DownloadRemoteFile struct {
	File    *sftp.File
	Total   int64
	Current int64
	Finish  bool
}

func (f *UpdateFile) Read(p []byte) (n int, err error) {
	n, err = f.File.Read(p)
	if f.Finish {
		return
	}
	f.Current += int64(n)
	var unit string = "M"
	var downloaded = f.Total / 1024 / 1024
	if downloaded == 0 {
		downloaded = f.Total / 1024
		unit = "K"
	}
	fmt.Printf("\r正在上传，共[%v]%v，进度：%.2f%%", downloaded, unit, float64(f.Current*10000/f.Total)/100)
	if f.Current == f.Total {
		f.Finish = true
		fmt.Printf("\r上传完成，共[%v]%v，进度：%.2f\n", downloaded, unit, float64(f.Current*10000/f.Total)/100)
	}
	return
}

func (f *DownloadRemoteFile) Read(p []byte) (n int, err error) {
	n, err = f.File.Read(p)
	if f.Finish {
		return
	}
	f.Current += int64(n)
	var unit string = "M"
	var downloaded = f.Total / 1024 / 1024
	if downloaded == 0 {
		downloaded = f.Total / 1024
		unit = "K"
	}
	fmt.Printf("\r正在下载，共[%v]%v，进度：%.2f%%", downloaded, unit, float64(f.Current*10000/f.Total)/100)
	if f.Current == f.Total {
		f.Finish = true
		fmt.Printf("\r下载完成，共[%v]%v，进度：%.2f\n", downloaded, unit, float64(f.Current*10000/f.Total)/100)
	}
	return
}

func (c Cli) Sftp(sourceFile string, targetFile string) string {
	if c.client == nil {
		if err := c.connect(); err != nil {
			Log.Fatal("connect err, %v", err)
			return ""
		}
	}
	sftpCli, err := sftp.NewClient(c.client)
	if err != nil {
		Log.Fatal("unable to start sftp subsytem: %v", err)
	}
	defer sftpCli.Close()

	f, err := os.OpenFile(sourceFile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		Log.Error("Open sourceFile error, %v", err)
		return err.Error()
	}
	defer f.Close()
	updateStat, _ := f.Stat()
	// 如果目标文件已存在，且大小与待上传文件大小一致，无需上传
	var remoteFile *sftp.File
	remoteFile, err = sftpCli.OpenFile(targetFile, os.O_RDONLY)
	locaSignature, _ := FileSignature(f)

	// 获取签名后，会关闭文件，需要重新打开
	f, _ = os.OpenFile(sourceFile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		Log.Debug("读取远程文件[%v]失败，需要重新上传，原因：%v\n", targetFile, err.Error())
	} else {
		remoteStat, _ := remoteFile.Stat()
		if remoteStat.Size() == updateStat.Size() {
			// 如果文件大小一样，且签名一样，无需下载
			remoteSignature := GetRemoteSignature(sftpCli, targetFile)
			if remoteSignature == locaSignature {
				Log.Trace("服务器已存在该文件[%v]，大小[%vM]，无需上传", remoteStat.Name(), remoteStat.Size()/1024/1024)
				return "exist"
			}
		} else {
			Log.Trace("服务器文件与本地文件大小不一致，需要上传")
		}
	}
	remoteFile, err = sftpCli.OpenFile(targetFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE)
	if err != nil {
		Log.Fatal("远程文件[%v]写入失败，检查服务器路径是否存在。错误： %v", targetFile, err)
		return ""
	}
	defer remoteFile.Close()

	t1 := time.Now()

	updateFile := &UpdateFile{
		File:   f,
		Total:  updateStat.Size(),
		Finish: false,
	}
	_, err = io.Copy(remoteFile, updateFile)
	if err != nil {
		Log.Fatal("copy file error, %v", err)
		return ""
	}

	Log.Debug("上传文件[%v]完毕，用时 %.2f 秒", updateStat.Name(), time.Since(t1).Seconds())
	remoteMd5File, _ := sftpCli.OpenFile(targetFile+".md5", os.O_WRONLY|os.O_TRUNC|os.O_CREATE)
	defer remoteMd5File.Close()
	_, err = remoteMd5File.Write([]byte(locaSignature))
	if err != nil {
		Log.Error("上传文件[%v]签名失败，错误：%v", updateStat.Name(), err.Error())
	} else {
		Log.Debug("上传文件[%v]签名成功：%v", updateStat.Name(), locaSignature)
	}
	return "success"
}

func GetRemoteSignature(c *sftp.Client, targetFile string) string {
	md5File, err := c.OpenFile(targetFile+".md5", os.O_RDONLY)
	if err != nil {
		Log.Debug("没有获取到远程文件的签名，文件：%v\n", targetFile)
		return "not exist"
	}
	defer md5File.Close()
	buf, _ := ioutil.ReadAll(md5File)
	return string(buf)
}

func (c Cli) SftpDown(targetFile string, locPath string) string {
	if c.client == nil {
		if err := c.connect(); err != nil {
			Log.Fatal("connect err, %v", err)
			return ""
		}
	}
	sftpCli, err := sftp.NewClient(c.client)
	if err != nil {
		Log.Fatal("unable to start sftp subsytem, %v", err)
	}
	defer sftpCli.Close()

	var remoteFile *sftp.File
	remoteFile, err = sftpCli.OpenFile(targetFile, os.O_RDONLY)
	if err != nil {
		Log.Error("读取文件[%v]失败，不能下载，error： %v", targetFile, err)
		return err.Error()
	}

	remoteStat, _ := remoteFile.Stat()
	remoteFileName := remoteStat.Name()
	defer remoteFile.Close()

	localFilePath := locPath + remoteFileName
	var locaSignature string
	remoteSignature := GetRemoteSignature(sftpCli, targetFile)
	if PathExists(localFilePath) {
		// 如果本地文件已存在，判断文件大小，如果与远程文件大小一致，无需下载
		localFile, _ := os.Open(localFilePath)
		localFileStat, _ := localFile.Stat()
		if localFileStat.Size() == remoteStat.Size() {
			// 如果文件大小一样，且签名一样，无需下载
			locaSignature, _ = FileSignature(localFile)
			if remoteSignature == locaSignature {
				Log.Trace("本地文件[%v]与远程文件相同，无需下载", remoteFileName)
				return "exist"
			}
		}
	}
	f, err := os.Create(localFilePath)
	if err != nil {
		Log.Fatal("Open sourceFile error, %v", err)
		return ""
	}
	defer f.Close()

	t1 := time.Now()

	downRemoteFile := &DownloadRemoteFile{
		File:   remoteFile,
		Total:  remoteStat.Size(),
		Finish: false,
	}
	Log.Debug("开始下载远程文件[%v]", remoteFileName)
	_, err = io.Copy(f, downRemoteFile)
	if err != nil {
		Log.Fatal("copy file error, %v", err)
		return ""
	}

	Log.Debug("远程文件[%v]下载完毕，用时 %.2f 秒", remoteFileName, time.Since(t1).Seconds())
	return "success"
}

//连接
func (c *Cli) connect() error {
	config := ssh.ClientConfig{
		User: c.Username,
		Auth: []ssh.AuthMethod{ssh.Password(c.Password)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: 10 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", c.IP, c.Port)
	sshClient, err := ssh.Dial("tcp", addr, &config)
	if err != nil {
		return err
	}
	c.client = sshClient
	return nil
}
