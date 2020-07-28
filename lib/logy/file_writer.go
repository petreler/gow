package logy

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

type ByType int

// GetFileFormat 获取格式
func (b ByType) GetFileFormat() string {
	return formats[b]
}

// SetFileFormat 设置文件格式
func SetFileFormat(t ByType, format string) {
	formats[t] = format
}

const (
	Day ByType = iota
	Hour
	Month
)

var (
	//formats 文件格式map
	formats = map[ByType]string{
		Day:   "2006-01-12",
		Hour:  "2006-01-02-15",
		Month: "2006-01",
	}
)

type FileOptions struct {
	Dir    string
	ByType ByType
	Loc    *time.Location
}

type Files struct {
	FileOptions
	file       *os.File
	lastFormat string
	mu         sync.Mutex
}

// NewFileWriter 设置写文件参数
// 	w:=logy.NewFileWriter(logy.FileOptions{
//		ByType:log.Day,
//		Dir:"./logs",
//	})
//  logy.Std.SetOutPut(w)
func NewFileWriter(opts ...FileOptions) *Files {
	opt := prepareFileOption(opts)
	return &Files{
		FileOptions: opt,
	}
}

func (f *Files) getFile() (*os.File, error) {
	var err error
	t := time.Now().In(f.Loc)
	if f.file == nil {
		f.lastFormat = t.Format(f.ByType.GetFileFormat())
		f.file, err = os.OpenFile(filepath.Join(f.Dir, f.lastFormat+".log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		return f.file, err
	}
	if f.lastFormat != t.Format(f.ByType.GetFileFormat()) {
		f.file.Close()
		f.lastFormat = t.Format(f.ByType.GetFileFormat())
		f.file, err = os.OpenFile(filepath.Join(f.Dir, f.lastFormat+".log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		return f.file, err
	}
	return f.file, nil
}

func (f *Files) Write(bs []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	w, err := f.getFile()
	if err != nil {
		return 0, err
	}
	return w.Write(bs)
}

func (f *Files) Close() {
	if f.file != nil {
		f.file.Close()
		f.file = nil
	}
	f.lastFormat = ""
}

func prepareFileOption(opts []FileOptions) FileOptions {
	var opt FileOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	if opt.Dir == "" {
		opt.Dir = "./"
	}
	err := os.MkdirAll(opt.Dir, os.ModePerm)
	if err != nil {
		panic(err.Error())
	}

	if opt.Loc == nil {
		opt.Loc = time.Local
	}
	return opt
}
