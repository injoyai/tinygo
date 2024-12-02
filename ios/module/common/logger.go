package common

import (
	"encoding/hex"
	"github.com/injoyai/tinygo/logs"
)

type Logger interface {
	Readln(prefix string, p []byte)
	Writeln(prefix string, p []byte)
	Infof(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	SetEncode(f func(p []byte) string)
	WithUTF8()
	WithHEX()
	Debug(b ...bool)
	SetLevel(level logs.Level)
}

func NewLogger() *logger {
	return &logger{
		err:    logs.New("错误").SetSelfLevel(logs.LevelError),
		info:   logs.New("信息").SetSelfLevel(logs.LevelInfo),
		read:   logs.New("读取").SetSelfLevel(logs.LevelRead),
		write:  logs.New("写入").SetSelfLevel(logs.LevelWrite),
		encode: func(p []byte) string { return string(p) },
		debug:  true,
	}
}

type logger struct {
	write  *logs.Log
	read   *logs.Log
	info   *logs.Log
	err    *logs.Log
	encode func(p []byte) string
	debug  bool
}

func (this *logger) Debug(b ...bool) {
	this.debug = len(b) == 0 || b[0]
}

func (this *logger) SetLevel(level logs.Level) {
	this.write.SetLevel(level)
	this.read.SetLevel(level)
	this.info.SetLevel(level)
	this.err.SetLevel(level)
}

func (this *logger) WithUTF8() {
	this.SetEncode(func(p []byte) string { return string(p) })
}

func (this *logger) WithHEX() {
	this.SetEncode(func(p []byte) string { return hex.EncodeToString(p) })
}

func (this *logger) SetEncode(f func(p []byte) string) {
	this.encode = f
}

func (this *logger) Readln(prefix string, p []byte) {
	if !this.debug {
		return
	}
	s := this.encode(p)
	this.read.Printf("%s%s\n", prefix, s)
}

func (this *logger) Writeln(prefix string, p []byte) {
	if !this.debug {
		return
	}
	s := this.encode(p)
	this.write.Printf("%s%s\n", prefix, s)
}

func (this *logger) Infof(format string, v ...interface{}) {
	if !this.debug {
		return
	}
	this.info.Printf(format, v...)
}

func (this *logger) Errorf(format string, v ...interface{}) {
	if !this.debug {
		return
	}
	this.err.Printf(format, v...)
}
