package oss

import (
	"fmt"
	"github.com/injoyai/tinygo/conv"
	"time"
)

// 64位最大只能到15.999EB
var mapSizeUnit = []string{
	"B",
	"KB",
	"MB",
	"GB",
	"TB",
	"PB",
	"EB", //64位最大单位
	"ZB",
	"YB",
	"BB",
	"NB",
	"DB",
	"CB",
	"XB",
}

// SizeUnit 字节数量和单位 例 15.8,"MB"
// 64位最大值是 18446744073709551616 = 15.999EB
func SizeUnit(b int64) (float64, string) {
	return Volume(b).SizeUnit()
}

// SizeString 字节数量字符表现方式,例 15.8MB, 会四舍五入
func SizeString(b int64, decimal ...int) string {
	size, unit := SizeUnit(b)
	d := conv.GetDefaultInt(1, decimal...)
	return fmt.Sprintf(fmt.Sprintf("%%.%df%%s", d), size, unit)
}

// SizeSpeed 每秒速度 例15.8MB/s
func SizeSpeed(b int64, sub time.Duration, decimal ...int) string {
	size, unit := SizeUnit(b)
	spend := size / sub.Seconds()
	d := conv.GetDefaultInt(1, decimal...)
	return fmt.Sprintf(fmt.Sprintf("%%.%df%%s/s", d), spend, unit)
}
