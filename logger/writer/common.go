package writer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ErrWriteTimeout 写超时错误
var ErrWriteTimeout = errors.New("write timeout")

func log2Stderr(format string, vs ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	prefix := strings.Join([]string{
		time.Now().Format("20060102 15:04:05.999999999"),
		" ",
		"writer/",
		filepath.Base(file),
		":",
		strconv.Itoa(line),
	}, "")

	fmt.Fprintf(os.Stderr, prefix+" "+format, vs...)
}

// exists 判断文件或目录是否存在
func exists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil {
		return true
	}
	return !os.IsNotExist(err)
}

func isSameFile(f1, f2 string) bool {
	s1, e1 := os.Stat(f1)
	s2, e2 := os.Stat(f2)
	if e1 != nil || e2 != nil {
		return false
	}
	return os.SameFile(s1, s2)
}

var isDebug = false

func init() {
	isDebug = os.Getenv("extension_writer") == "debug"
}

// keepDirExists 保持文件目录存在，若不存在则创建
// 若地址是文件不是目录，会将文件重命名，然后创建目录
func keepDirExists(dir string) error {
	info, errStat := os.Stat(dir)
	if errStat == nil && info.IsDir() {
		return nil
	}

	// 文件存在，但是不是目录,将其重命名
	if errStat == nil {
		newName := dir + "_not_dir_" + strconv.FormatInt(time.Now().UnixNano(), 10)
		if errRename := os.Rename(dir, newName); errRename != nil {
			if !os.IsNotExist(errRename) {
				return errRename
			}
		}
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		if os.IsExist(err) {
			return nil
		}
		return err
	}
	return nil
}

var nowFunc = time.Now
