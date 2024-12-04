package oss

import (
	"encoding/csv"
	"github.com/injoyai/tinygo/conv"
	"os"
)

type CSVFile struct {
	*os.File
	*csv.Writer
}

func (this *CSVFile) Write(v ...interface{}) (err error) {
	switch len(v) {
	case 0:
	case 1:
		err = this.Writer.Write(conv.Strings(v[0]))
	default:
		err = this.Writer.Write(conv.Strings(v))
	}
	return
}

func (this *CSVFile) WriteFlush(v ...interface{}) error {
	defer this.Flush()
	return this.Write(v...)
}
