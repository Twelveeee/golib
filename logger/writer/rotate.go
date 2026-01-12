package writer

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Twelveeee/golib/logger/fileclean"
)

// RotateOption NewRotate的参数
type RotateOption struct {
	// 用于控制当前写入那个文件，会定期变化
	// 如每小时返回一个新的文件名
	FileProducer RotateProducer

	// 用于控制文件落盘的间隔，若超过该时间，将强制性Flush
	// 默认为0，不定期Flush
	// 由于Rotate Writer 用到BufWriter，所以若写入量很少，内容落盘将出现延迟
	FlushDuration time.Duration

	// CheckDuration 检查文件是否存在的时间间隔
	// 用于处理 文件被删除或者改名的情况
	// 如间隔1秒检查，默认为0，不检查
	CheckDuration time.Duration

	// 保留最多日志文件数，默认为0,不清理
	MaxFileNum int
}

// Check 检查参数是否正确
func (ro *RotateOption) Check() error {
	if ro.FileProducer == nil {
		return errors.New("fileProducer required")
	}
	info := ro.FileProducer.Get()
	if info.FilePath == "" {
		return errors.New("fileProducer.Get().FilePath is empty")
	}
	return nil
}

// NewRotate 创建一个具有切换文件名的文件writer
func NewRotate(opt *RotateOption) (io.WriteCloser, error) {
	if opt == nil {
		return nil, errors.New("rotateOption is nil")
	}
	if err := opt.Check(); err != nil {
		return nil, err
	}

	w := &rotateWriter{
		opt: opt,
	}
	if err := w.init(); err != nil {
		_ = w.Close()
		return nil, err
	}
	return w, nil
}

type rotateWriter struct {
	outFile     *os.File
	outFileInfo os.FileInfo

	bufFile *bufio.Writer

	mu sync.Mutex
	// flush 的间隔，若超过该时长没有进行Flush,将触发刷新
	lastFlush time.Time

	opt *RotateOption

	onCloseFuncs []func()

	// 清理文件时的延迟时间，避免集中清理
	cleanDelay func() time.Duration
}

func (f *rotateWriter) init() error {
	opt := f.opt
	rp := opt.FileProducer
	if err := f.checkOpened(rp.Get()); err != nil {
		return err
	}

	rp.RegisterCallBack(func(info RotateInfo) {
		_ = f.checkOpened(info)
	})

	f.onClose(func() {
		_ = rp.Stop()
	})

	if f.cleanDelay == nil {
		f.cleanDelay = func() time.Duration {
			return time.Second * time.Duration(5+rand.Intn(60))
		}
	}

	// MaxFileNum >0 表示需要进行文件清理
	if opt.MaxFileNum > 0 {
		rp.RegisterCallBack(func(info RotateInfo) {
			delay := f.cleanDelay()
			if delay > 0 {
				// 清理文件可以延迟一些，这样可以避免同一个机器上多个不同的应用
				// 在同一瞬间清理照成 io 压力大
				<-time.After(delay)
			}
			f.clean()
		})

		f.clean() // 启动阶段进行一次清理
	}

	// 定期刷新文件落盘
	if opt.FlushDuration > 0 {
		flushTicker := time.NewTicker(opt.FlushDuration)

		f.onClose(func() {
			flushTicker.Stop()
		})

		go func() {
			for range flushTicker.C {
				f.checkFlush(opt.FlushDuration)
			}
		}()
	}

	// 定期检查文件是否存在
	if opt.CheckDuration > 0 {
		checkTicker := time.NewTicker(opt.CheckDuration)
		f.onClose(func() {
			checkTicker.Stop()
		})
		go func() {
			for range checkTicker.C {
				if err := f.checkOpened(rp.Get()); err != nil {
					log2Stderr("checkDuration has error: %v\n", err)
				}
			}
		}()
	}

	return nil
}

func (f *rotateWriter) clean() {
	rawName := f.opt.FileProducer.Get().RawName
	files, err := fileclean.FindFiles(rawName, f.opt.MaxFileNum)
	if err != nil {
		log2Stderr("[rotate.clean] FindFiles(%q) has error:%v\n", rawName, err)
		return
	}
	for _, name := range files {
		err := os.Remove(name)
		log2Stderr("[rotate.clean] remove file %q for %q, err=%v\n", name, rawName, err)
	}
}

// onClose 注册在close前执行的回调方法
func (f *rotateWriter) onClose(fn func()) {
	f.onCloseFuncs = append(f.onCloseFuncs, fn)
}

// checkOpened 检查文件是否打开，若没有打开将打开文件
// 若当前期望写入的文件名和之前已打开的文件不一致，将先关闭，然后打开新的文件句柄
func (f *rotateWriter) checkOpened(info RotateInfo) (errResult error) {
	f.mu.Lock()
	fileExists := f.outFileExists(info.FilePath)
	f.mu.Unlock()

	defer func() {
		if errResult != nil {
			log2Stderr("checkOpened has error: %s\n", errResult.Error())
		}
	}()

	if !fileExists {
		dir := filepath.Dir(info.FilePath)
		if err := keepDirExists(dir); err != nil {
			return err
		}
	}

	needNew := true

	f.mu.Lock()
	defer f.mu.Unlock()

	if f.outFile != nil && fileExists {
		needNew = false
	}

	if needNew {
		if f.outFile != nil {
			errFlush := f.bufFile.Flush()
			errClose := f.outFile.Close()

			if errFlush != nil || errClose != nil {
				log2Stderr("close old file has error, flush=%v, close=%v\n", errFlush, errClose)
			}
		}

		logFile, errOpen := os.OpenFile(info.FilePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
		if errOpen != nil {
			return fmt.Errorf("os.OpenFile(%q,xx,0644) has error:%w", info.FilePath, errOpen)
		}

		{
			fileStat, errStat := logFile.Stat()
			if errStat != nil {
				return fmt.Errorf("read %q's stat error: %w", info.FilePath, errStat)
			}
			f.outFileInfo = fileStat
		}

		f.outFile = logFile
		f.bufFile = bufio.NewWriter(f.outFile)
	}

	return f.checkSymlink(info)
}

// checkSymlink 检查文件软连接是否存在
func (f *rotateWriter) checkSymlink(info RotateInfo) error {
	return checkSymlink(info)
}

// outFileExists 判断outFile存在，并且文件Stat没有变化
func (f *rotateWriter) outFileExists(outFile string) bool {
	if !exists(outFile) {
		return false
	}
	info, err := os.Stat(outFile)
	if err != nil {
		return false
	}
	return os.SameFile(info, f.outFileInfo)
}

func (f *rotateWriter) Write(p []byte) (n int, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.bufFile == nil {
		return 0, io.ErrClosedPipe
	}

	n, err = f.bufFile.Write(p)

	if f.bufFile.Buffered() == 0 {
		f.lastFlush = time.Now()
	}

	return n, err
}

// Flush 文件内容刷新落盘
func (f *rotateWriter) Flush() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.bufFile == nil {
		return nil
	}
	return f.bufFile.Flush()
}

func (f *rotateWriter) checkFlush(dur time.Duration) {
	f.mu.Lock()
	lastFlush := f.lastFlush
	f.mu.Unlock()

	if lastFlush.IsZero() || time.Since(lastFlush) >= dur {
		_ = f.Flush()
	}
}

// Close 关闭writer
func (f *rotateWriter) Close() error {
	for _, fn := range f.onCloseFuncs {
		fn()
	}
	var err1, err2 error
	f.mu.Lock()
	if f.bufFile != nil {
		err1 = f.bufFile.Flush()
	}
	if f.outFile != nil {
		err2 = f.outFile.Close()
	}
	f.outFile = nil
	f.bufFile = nil
	f.mu.Unlock()

	if err1 == nil && err2 == nil {
		return nil
	}
	return fmt.Errorf("flush:%w, close:%v", err1, err2)
}

var _ io.WriteCloser = (*rotateWriter)(nil)

// checkSymlink 检查并保持软连正确
func checkSymlink(info RotateInfo) error {
	if !info.NeedSymlink() {
		return nil
	}

	if !exists(info.FilePath) {
		return fmt.Errorf("file=%q not exists", info.FilePath)
	}

	symDir := filepath.Dir(info.Symlink)
	if !exists(symDir) {
		if err := os.MkdirAll(symDir, 0755); err != nil && !os.IsExist(err) {
			return err
		}
	}

	// 若软连接对应的文件已存在，需要进行预处理
	// 1.若是软连，直接删除
	// 2.若判断软连的时候出错，发现已不存在，什么都不做
	// 3.其他情况（不是软连，可能是文件），将其重命名
	if exists(info.Symlink) {
		// 判断 os.IsNotExist 是为了更好的兼容一个文件同时被多个writer 或者多个程序切分
		if _, err := os.Readlink(info.Symlink); err == nil {
			// 若已经是当前文件的软连
			if isSameFile(info.Symlink, info.FilePath) {
				return nil
			}

			if isDebug {
				log.Println("remove Symlink", info.Symlink)
			}

			// 若是软连接,但是不是当前文件，则删除
			if errRm := os.Remove(info.Symlink); errRm != nil && !os.IsNotExist(errRm) {
				return fmt.Errorf("os.Remove %q has error: %w", info.Symlink, errRm)
			}
		} else if os.IsNotExist(err) {
			// do nothing
		} else {
			// 其他情况，则重命名
			newName := fmt.Sprintf("%s.old.%s", info.Symlink, time.Now().Format("20060102150405"))
			if errRe := os.Rename(info.Symlink, newName); errRe != nil && !os.IsNotExist(errRe) {
				return fmt.Errorf("os.Rename(%q,%q) has error: %w", info.Symlink, newName, errRe)
			}
		}
	}

	// 创建软连接
	{
		name := info.FilePath
		// 获取相对路径，这样可以保证在不同的运行环境下，都能正常的读取到正确的原始路径
		if relPath, err := filepath.Rel(symDir, info.FilePath); err == nil {
			name = relPath
		}

		if isDebug {
			// 这个在单测里会用到
			log.Println("create Symlink", name, info.Symlink)
		}

		// 当出现 os.IsExist(errSl) 时，可能其他的rotator 已经运行过了
		// 比如外部的日志切分程序，或者是对同一个文件同时有多个writer
		if errSl := os.Symlink(name, info.Symlink); errSl != nil && !os.IsExist(errSl) {
			return fmt.Errorf("os.Symlink(%q,%q) has error:%w", name, info.Symlink, errSl)
		}
	}
	return nil
}
