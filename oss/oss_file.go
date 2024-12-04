package oss

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"github.com/injoyai/tinygo/conv"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	defaultPerm = 0777
)

// Read 读取文件
func Read(filename string) ([]byte, error) {
	return ReadBytes(filename)
}

// ReadReader 读取reader
func ReadReader(r io.Reader) ([]byte, error) {
	return ioutil.ReadAll(r)
}

// ReadBytes 读取文件内容
func ReadBytes(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

// ReadString 读取文件
func ReadString(filename string) (string, error) {
	bs, err := ReadBytes(filename)
	return string(bs), err
}

// ReadHEX 读取文件
func ReadHEX(filename string) (string, error) {
	bs, err := ReadBytes(filename)
	return hex.EncodeToString(bs), err
}

// ReadBase64 读取文件
func ReadBase64(filename string) (string, error) {
	bs, err := ReadBytes(filename)
	return base64.StdEncoding.EncodeToString(bs), err
}

// NewFile 新建文件
func NewFile(filename string) (*os.File, error) {
	return os.Create(filename)
}

// OpenFile 打开文件
func OpenFile(filename string) (*os.File, error) {
	return os.Open(filename)
}

// New 新建文件,会覆盖
func New(filename string, v ...interface{}) error {
	if len(v) == 0 {
		return os.MkdirAll(filename, defaultPerm)
	}
	dir, name := filepath.Split(filename)
	if len(dir) > 0 {
		if err := os.MkdirAll(dir, defaultPerm); err != nil {
			return err
		}
	}
	if len(name) == 0 {
		return nil
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, k := range v {
		switch r := k.(type) {
		case nil:
		case io.Reader:
			if _, err = io.Copy(f, r); err != nil {
				return err
			}
		default:
			if _, err = f.Write(conv.Bytes(r)); err != nil {
				return err
			}
		}
	}
	return nil
}

// NewNotExist 如果不存在,则新建
func NewNotExist(filename string, v ...interface{}) error {
	if !Exists(filename) {
		return New(filename, v...)
	}
	return nil
}

// OpenCSV 新建或者打开csv文件
func OpenCSV(filename string, initStr ...interface{}) (*CSVFile, error) {
	if !Exists(filename) {
		f, err := os.Create(filename)
		if err != nil {
			return nil, err
		}
		ff := &CSVFile{
			File:   f,
			Writer: csv.NewWriter(f),
		}

		//写入utf-8 编码
		if _, err = f.WriteString("\xEF\xBB\xBF"); err == nil {
			//写入预设值,例如标题
			err = ff.Write(initStr...)
		}

		if err != nil {
			f.Close()
		}
		return ff, err
	}

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, defaultPerm)
	return &CSVFile{
		File:   f,
		Writer: csv.NewWriter(f),
	}, err
}

func WithOpenCSV(filename string, fn func(f *CSVFile), initStr ...interface{}) error {
	f, err := OpenCSV(filename, initStr...)
	if err != nil {
		return err
	}
	defer f.Close()
	fn(f)
	return nil
}

func OpenAppend(filename string) (*os.File, error) {
	return os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, defaultPerm)
}

func WithOpenAppend(filename string, fn func(f *os.File) error) error {
	f, err := OpenAppend(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return fn(f)
}

// OpenFunc 打开文件,并执行函数
func OpenFunc(filename string, fn func(f *os.File) error) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return fn(f)
}

func WithOpen(filename string, fn func(f *os.File) error) error {
	return OpenFunc(filename, fn)
}

// WithCopyTo 打开文件,并写入到io.Writer
func WithCopyTo(filename string, writer io.Writer) error {
	return OpenFunc(filename, func(f *os.File) error {
		_, err := io.Copy(writer, f)
		return err
	})
}

// WithCopyFrom 打开文件,并从io.Reader写入
func WithCopyFrom(filename string, reader io.Reader) error {
	return OpenFunc(filename, func(f *os.File) error {
		_, err := io.Copy(f, reader)
		return err
	})
}
