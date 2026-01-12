package fileclean

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// FindFiles 查找要清理的文件
// 按照文件创建时间排序，先创建的先返回
// 查找的文件名匹配的内容只能包含一个".",而且只能是 《.数字》 结尾
// keep 参数控制剩余文件数
func FindFiles(prefixName string, keep int) ([]string, error) {
	pattern := prefixName + ".*"
	matches, errGlob := filepath.Glob(pattern)
	if errGlob != nil {
		return nil, errGlob
	}
	if len(matches) <= keep {
		return nil, nil
	}
	// 原始的文件名 如 ral-worker.log
	baseName := filepath.Base(prefixName)

	infos := make([]os.FileInfo, 0, len(matches))
	for i := 0; i < len(matches); i++ {
		// name eg: ral-worker.log.2020123115
		name := matches[i]
		info, errStat := os.Stat(name)
		if errStat != nil {
			if os.IsNotExist(errStat) {
				continue
			}
			// 其他情况，可以打印一条日志
			log.Printf("os.Stat(%q) has error:%v\n", name, errStat)
			continue
		}

		if info.IsDir() {
			continue
		}

		if !isFileNameMatch(baseName, filepath.Base(name)) {
			continue
		}

		infos = append(infos, info)
	}

	// 按照文件的创建时间排序
	sort.Slice(infos, func(i, j int) bool {
		a := infos[i]
		b := infos[j]
		return ctime(a) < ctime(b)
	})

	var result []string
	dir := filepath.Dir(pattern)
	for i := 0; i < len(infos)-keep; i++ {
		name := filepath.Join(dir, infos[i].Name())
		result = append(result, name)
	}
	return result, nil
}

var extReg = regexp.MustCompile(`\.\d+`)

// isFileNameMatch 判断文件名是否含有特定的前缀
// 除了前缀部分后,其他部分只能是 .XXX 格式，同时XXX不能包含"."
func isFileNameMatch(prefix string, name string) bool {
	if !strings.HasPrefix(name, prefix) {
		return false
	}

	// 文件后缀， eg： .2020123115、.wf.2020123115
	extName := name[len(prefix):]
	if len(extName) == 0 || extName[0] != '.' {
		return false
	}

	// 若包含多个"." 说明不是当前任务查找的文件
	// 比如
	// 1.输入 ral-worker.log 期望 找到文件 ral-worker.log.2020123115
	// 而不期望找到文件 ral-worker.log.wf.2020123115
	// 2.输入 ral-worker.log.wf 期望找到文件 ral-worker.log.wf.2020123115
	if strings.Count(extName, ".") > 1 {
		return false
	}
	return extReg.MatchString(extName)
}
