package logs

import (
	"time"
	"unsafe"
)

type Formatter interface {
	Formatter(l *Log, bs []byte) []byte
}

var (
	// DefaultFormatter 默认格式化可修改
	DefaultFormatter Formatter = FormatFunc(timeFormatter)
)

type FormatFunc func(l *Log, bs []byte) []byte

func (this FormatFunc) Formatter(l *Log, bs []byte) []byte {
	return this(l, bs)
}

func timeFormatter(l *Log, bs []byte) []byte {
	s := time.Now().Format("15:04:05") + " [" + l.Tag + "] " + string(bs)
	return *(*[]byte)(unsafe.Pointer(&s))
}
