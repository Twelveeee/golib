// go:build !linux && !windows
package fileclean

import (
	"os"
	"syscall"
)

func ctime(info os.FileInfo) int64 {
	stat := info.Sys().(*syscall.Stat_t)
	return stat.Ctimespec.Nano()
}
