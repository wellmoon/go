package utils

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"

	Log "github.com/wellmoon/go/logger"
)

func Zip(srcFile string) error {
	Log.Trace("压缩文件：%v", srcFile)
	destZip := srcFile + ".zip"
	zipfile, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	_, root := filepath.Split(srcFile)
	//sep := string(os.PathSeparator)
	filepath.Walk(srcFile, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if strings.Contains(path, ".git") {
			return nil
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// 压缩包的根目录为待压缩文件夹，不包含其上级的所有目录
		if path == srcFile {
			header.Name = root
		} else {
			sub := strings.Split(path, root)[1]
			// 如果程序运行在windows下，路径分隔符为\，需要替换为 /
			// 需要在linux下解压缩，路径中所有的分隔符要替换为/
			linuxSub := strings.ReplaceAll(sub, `\`, `/`)
			header.Name = root + linuxSub
		}

		if info.IsDir() {
			header.Name = header.Name + "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, _ = io.Copy(writer, file)
		}
		return err
	})

	return err
}

func Unzip(zipFile string, destDir string) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		fpath := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return err
			}

			inFile, err := f.Open()
			if err != nil {
				return err
			}
			defer inFile.Close()

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
