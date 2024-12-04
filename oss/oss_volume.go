package oss

import (
	"fmt"
	"github.com/injoyai/tinygo/conv"
	"math"
	"strings"
)

const (
	B  = 1
	KB = B * 1024
	MB = KB * 1024
	GB = MB * 1024
	TB = GB * 1024
	PB = TB * 1024
	EB = PB * 1024
)

type Volume uint64

func (this Volume) Uint64() uint64 {
	return uint64(this)
}

func (this Volume) String() string {
	size, unit := this.SizeUnit()
	return fmt.Sprintf("%v%s", size, unit)
}

func (this Volume) SizeUnit() (float64, string) {
	i := 0
	//先用uint64进行循环除以1024,float64的指数比uint64小
	//当值比float64的最大值小的时候,能转成float64时,使用float64进行除以1024
	for ; float64(this) < 0; i++ {
		this = this / 1024
	}
	f := float64(this)
	for ; f >= 1024*1024; i++ {
		f = f / 1024
	}
	if f >= 1024 {
		return f / 1024, mapSizeUnit[i+1]
	}
	return f, mapSizeUnit[i]
}

// ParseVolume 解析体积,
func ParseVolume(s string) Volume {
	total := Volume(0)
	size := ""
	unit := ""
	hasUnit := false

	add := func() {
		if hasUnit {
			for i, u := range mapSizeUnit {
				if strings.ToUpper(unit) == u {
					total += Volume(conv.Float64(size) * math.Pow(1024, float64(i)))
				}
			}
			hasUnit = false
			size = ""
			unit = ""
		}
	}

	for _, v := range s {
		if (v >= '0' && v <= '9') || v == '.' {
			add()
			size += string(v)
		} else {
			unit += string(v)
			hasUnit = true
		}
	}
	add()

	return total
}
