package logs

import (
	"fmt"
	"io"
	"time"
)

var (
	DefaultTrace = New("TRACE").SetSelfLevel(LevelTrace)
	DefaultDebug = New("DEBUG").SetSelfLevel(LevelDebug)
	DefaultWrite = New("WRITE").SetSelfLevel(LevelWrite)
	DefaultRead  = New("READ ").SetSelfLevel(LevelRead)
	DefaultInfo  = New("INFO ").SetSelfLevel(LevelInfo)
	DefaultWarn  = New("WARN ").SetSelfLevel(LevelWarn)
	DefaultErr   = New("ERROR").SetSelfLevel(LevelError)
)

// SetWriter 覆盖io.Writer
func SetWriter(w io.Writer) {
	DefaultTrace.SetWriter(w)
	DefaultDebug.SetWriter(w)
	DefaultWrite.SetWriter(w)
	DefaultRead.SetWriter(w)
	DefaultInfo.SetWriter(w)
	DefaultWarn.SetWriter(w)
	DefaultErr.SetWriter(w)
}

// SetLevel 设置日志等级
func SetLevel(level Level) {
	DefaultTrace.SetLevel(level)
	DefaultDebug.SetLevel(level)
	DefaultWrite.SetLevel(level)
	DefaultRead.SetLevel(level)
	DefaultInfo.SetLevel(level)
	DefaultWarn.SetLevel(level)
	DefaultErr.SetLevel(level)
}

// PrintErr 打印错误,有错误才打印
func PrintErr(err error) bool {
	if err != nil {
		DefaultErr.Println(err.Error())
	}
	return err != nil
}

// Spend 记录耗时,使用方式 defer Spend()()
func Spend(prefix ...interface{}) func() {
	now := time.Now()
	return func() {
		DefaultDebug.Println(fmt.Sprint(prefix...) + time.Now().Sub(now).String())
	}
}

// Trace 预设追溯
// [追溯] 2022/01/08 10:44:02 init_test.go:10:
func Trace(s ...interface{}) (int, error) {
	return DefaultTrace.Println(s...)
}

// Tracef 预设追溯
// [追溯] 2022/01/08 10:44:02 init_test.go:10:
func Tracef(format string, s ...interface{}) (int, error) {
	return DefaultTrace.Printf(format, s...)
}

// Debug 预设调试
// [调试] 2022/01/08 10:44:02 init_test.go:10:
func Debug(s ...interface{}) (int, error) {
	return DefaultDebug.Println(s...)
}

// Debugf 预设调试
// [调试] 2022/01/08 10:44:02 init_test.go:10:
func Debugf(format string, s ...interface{}) (int, error) {
	return DefaultDebug.Printf(format, s...)
}

// Read 预设读取
// [读取] 2022/01/08 10:44:02 init_test.go:10:
func Read(s ...interface{}) (int, error) {
	return DefaultRead.Println(s...)
}

// Readf 预设读取
// [读取] 2022/01/08 10:44:02 init_test.go:10:
func Readf(format string, s ...interface{}) (int, error) {
	return DefaultRead.Printf(format, s...)
}

// Write 预设写入
// [写入] 2022/01/08 10:44:02 init_test.go:10:
func Write(s ...interface{}) (int, error) {
	return DefaultWrite.Println(s...)
}

// Writef 预设写入
// [写入] 2022/01/08 10:44:02 init_test.go:10:
func Writef(format string, s ...interface{}) (int, error) {
	return DefaultWrite.Printf(format, s...)
}

// Info 预设信息
// [信息] 2022/01/08 10:44:02 init_test.go:10:
func Info(s ...interface{}) (int, error) {
	return DefaultInfo.Println(s...)
}

// Infof 预设信息
// [信息] 2022/01/08 10:44:02 init_test.go:10:
func Infof(format string, s ...interface{}) (int, error) {
	return DefaultInfo.Printf(format, s...)
}

// Warn 预设警告
// [警告] 2022/01/08 10:44:02 init_test.go:10:
func Warn(s ...interface{}) (int, error) {
	return DefaultWarn.Println(s...)
}

// Warnf 警告
func Warnf(format string, s ...interface{}) (int, error) {
	return DefaultWarn.Printf(format, s...)
}

// Err 预设错误
// [错误] 2022/01/08 10:44:02 init_test.go:10:
func Err(s ...interface{}) (int, error) {
	return DefaultErr.Println(s...)
}

// Errf 预设错误
// [错误] 2022/01/08 10:44:02 init_test.go:10:
func Errf(format string, s ...interface{}) (int, error) {
	return DefaultErr.Printf(format, s...)
}
