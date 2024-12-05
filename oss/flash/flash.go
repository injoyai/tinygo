package flash

import (
	"fmt"
	"github.com/injoyai/tinygo/logs"
	"unsafe"
)

/*
0x00000000 到 0x10000000：通常用作 程序代码，这是 Flash 存储的区域。
0x10000000 到 0x10100000：内置 Flash 存储（例如，用于存储程序和配置数据）。
0x20000000 到 0x20042000：内部 RAM 区域，共 264KB。

*/

const BufferSize uint32 = 256

const (
	PicoStart uint32 = 0x10180000 // 0x10000000
)

func New(start, size uint32) *Flash {
	return &Flash{
		Start: start,
		Size:  size,
	}
}

type Flash struct {
	Start uint32 //起始位置
	Size  uint32 //片区数量,1个片区等于256字节
}

func (this *Flash) Read(buf []byte) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	var slice *[BufferSize]byte
	bufMax := uint32(len(buf)) / BufferSize

	for i := uint32(0); i < this.Size && i <= bufMax; i++ {
		slice = (*[BufferSize]byte)(unsafe.Pointer(uintptr(this.Start + i*BufferSize)))
		logs.Debug(string(slice[:]))
		copy(buf[i*BufferSize:], slice[:])
	}

	return len(buf), nil
}

func (this *Flash) Write(buf []byte) (int, error) {
	var slice *[BufferSize]byte
	bufMax := uint32(len(buf)) / BufferSize

	for i := uint32(0); i < this.Size && i <= bufMax; i++ {
		slice = (*[BufferSize]byte)(unsafe.Pointer(uintptr(this.Start + i*BufferSize)))
		copy(slice[:], buf[i*BufferSize:])
	}
	return len(buf), nil
}
