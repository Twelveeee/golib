// go:build windows
package fileclean

import (
	"os"
)

func ctime(info os.FileInfo) int64 {
	return info.ModTime().UnixNano()
}
