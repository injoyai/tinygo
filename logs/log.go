package logs

import (
	"fmt"
	"io"
	"machine"
	"unsafe"
)

type Level uint8

const (
	LevelAll Level = iota
	LevelTrace
	LevelDebug
	LevelWrite
	LevelRead
	LevelInfo
	LevelWarn
	LevelError
	LevelNone Level = 255
)

func New(tag string) *Log {
	println()
	return &Log{
		Tag:       tag,
		Writer:    machine.DefaultUART,
		Formatter: DefaultFormatter,
	}
}

type Log struct {
	Tag       string    //日志标签
	SelfLevel Level     //自己的日志级别
	Level     Level     //日志级别
	Writer    io.Writer //输出目标
	Formatter Formatter //格式化输出
}

func (this *Log) SetWriter(writer io.Writer) {
	this.Writer = writer
}

func (this *Log) SetLevel(level Level) *Log {
	this.Level = level
	return this
}

func (this *Log) SetSelfLevel(level Level) *Log {
	this.SelfLevel = level
	return this
}

func (this *Log) Println(v ...any) (int, error) {
	if this.SelfLevel >= this.Level {
		if this.Formatter == nil {
			this.Formatter = DefaultFormatter
		}
		if this.Formatter == nil {
			this.Formatter = FormatFunc(timeFormatter)
		}
		s := fmt.Sprintln(v...)
		bs := this.Formatter.Formatter(this, *(*[]byte)(unsafe.Pointer(&s)))
		defer func() { bs = nil }()
		return this.Write(bs)
	}
	return 0, nil
}

func (this *Log) Printf(format string, v ...any) (int, error) {
	if this.SelfLevel >= this.Level {
		if this.Formatter == nil {
			this.Formatter = DefaultFormatter
		}
		if this.Formatter == nil {
			this.Formatter = FormatFunc(timeFormatter)
		}
		s := fmt.Sprintf(format, v...)
		bs := this.Formatter.Formatter(this, *(*[]byte)(unsafe.Pointer(&s)))
		defer func() { bs = nil }()
		return this.Write(bs)
	}
	return 0, nil
}

// Write 实现io.Writer
func (this *Log) Write(p []byte) (int, error) {
	if this.Writer == nil {
		return 0, nil
	}
	return this.Writer.Write(p)
}
