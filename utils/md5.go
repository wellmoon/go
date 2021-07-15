package utils

import (
	"bufio"
	"bytes"
	"container/list"
	"crypto/md5"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"sort"
	"sync"
)

// 大文件签名是，先切分成块，多线程签名
const blockNum int = 20

type File interface {
	Stat() (os.FileInfo, error)
	Seek(offset int64, whence int) (ret int64, err error)
	Read(b []byte) (n int, err error)
	ReadAt(b []byte, off int64) (n int, err error)
	Name() string
	Close() error
}

// 分解大文件，每个block生成md5并加入数组，排序后对整个数组md5
func Signature2(filename string) (string, error) {
	if info, err := os.Stat(filename); err != nil {
		return "", err
	} else if info.IsDir() {
		return "", nil
	}

	file, err := os.OpenFile(filename, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return "", err
	}
	defer file.Close()
	fileStat, _ := file.Stat()
	totalSize := fileStat.Size()
	if totalSize < 1024*1024*100 {
		// 文件小于100M，无需拆分
		return Md5File(file), nil
	}
	bufferSize := totalSize / int64(blockNum)

	fmt.Println("bufferSize:", bufferSize)
	var wg sync.WaitGroup
	var md5List = list.New()
	var bufList [blockNum + 1][]byte

	reader := bufio.NewReader(file)
	var buf []byte
	var idx = 0
	for {
		buf = make([]byte, bufferSize)
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		} else {
			bufList[idx] = buf[:n]
		}
		idx = idx + 1
	}
	for i := 0; i < len(bufList); i++ {
		wg.Add(1)
		go md5Block(md5List, bufList[i], &wg) // 多线程执行分段加密，提高速度
	}
	wg.Wait()
	var slice = make([]string, 0)
	for e := md5List.Front(); e != nil; e = e.Next() {
		str := fmt.Sprintf("%v", e.Value)
		slice = append(slice, str)
	}
	var b bytes.Buffer
	sort.Strings(sort.StringSlice(slice))
	for _, str := range slice {
		b.WriteString(str)
	}

	return Md5(b.Bytes()), nil
}

func Signature(filename string) (string, error) {
	if info, err := os.Stat(filename); err != nil {
		return "", err
	} else if info.IsDir() {
		return "", nil
	}

	file, err := os.OpenFile(filename, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return FileSignature(file)
}

func FileSignature(file File) (string, error) {
	defer file.Close()
	fileStat, _ := file.Stat()
	totalSize := fileStat.Size()
	if totalSize < 1024*1024*100 {
		// 文件小于100M，无需拆分
		res := Md5File(file)
		return res, nil
	}
	blockSize := totalSize / int64(blockNum)
	// var forNum int = blockNum
	// if totalSize%int64(blockNum) > 0 {
	// 	forNum = blockNum + 1
	// }
	var wg sync.WaitGroup
	m := make(map[int]string)

	// var bufList []string

	var readSize int64
	if blockSize > 100*1024*1024 {
		readSize = 100 * 1024 * 1024
	} else {
		readSize = blockSize
	}
	for i := 0; i < blockNum; i++ {
		wg.Add(1)
		go goRead(readSize, file, blockSize, i, m, &wg)
	}
	if totalSize%int64(blockNum) > 0 {
		file.Seek(int64(blockNum)*blockSize, 0)
		buf := make([]byte, totalSize-int64(blockNum)*blockSize)
		file.Read(buf)
		m[blockNum] = string(buf)
	}
	wg.Wait()
	//拿到key
	var keys []int
	for k := range m {
		keys = append(keys, k)
	}
	//对key排序
	sort.Ints(keys)
	//根据key从m中拿元素，就是按顺序拿了
	var b bytes.Buffer
	for _, k := range keys {
		b.WriteString(m[k])
	}
	return Md5(b.Bytes()), nil
}

func goRead(readSize int64, file File, blockSize int64, i int, m map[int]string, wg *sync.WaitGroup) {
	buf := make([]byte, readSize)
	defer wg.Done()
	_, err := file.ReadAt(buf, int64(i)*blockSize)
	if err != nil {
		fmt.Println(err, file.Name())
		return
	}
	res := Md5(buf)

	lock.Lock()
	m[i] = res
	lock.Unlock()

}

func Md5(buf []byte) string {
	hash := md5.New()
	_, err := hash.Write(buf)
	if err != nil {
		fmt.Printf("hash write err:%v", err)
	}

	checksum := fmt.Sprintf("%x", hash.Sum(nil))
	return checksum
}

var lock sync.RWMutex

func Md5File(file File) string {
	hash := md5.New()
	io.Copy(hash, file)
	checksum := fmt.Sprintf("%x", hash.Sum(nil))
	return checksum
}

func md5Block(list *list.List, buf []byte, wg *sync.WaitGroup) {
	str := Md5(buf)
	lock.Lock()
	list.PushBack(str)
	lock.Unlock()
	wg.Done()
}

func HashCode(s string) uint32 {
	return crc32.ChecksumIEEE([]byte(s))
}
