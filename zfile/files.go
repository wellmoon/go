package zfile

import (
	"os"

	Log "github.com/wellmoon/go/logger"
)

type FileUtil struct {
	Path string
}

func Create(filePath string) {
	f, err := os.Create(filePath)
	if err != nil {
		Log.Error("create file {} error : {}", filePath, err)
	}
	defer f.Close()
}
