package easy

import (
	"errors"
	"github.com/injoyai/tinygo/conv"
	"io"
)

const (
	Prefix = 0xFFF1
)

/*

包构成(大端):
帧头 2字节  0xFF01  01是版本v1
长度 2字节  最多支持65535字节
控制 1字节
数据 n字节
校验 1字节


*/

type Frame struct {
	Control uint8  //控制码
	Data    []byte //数据域
}

func (this *Frame) Bytes() []byte {
	length := len(this.Data)
	bs := make([]byte, length+6)
	bs[0] = uint8(Prefix >> 8)
	bs[1] = uint8(Prefix % 256)
	bs[2] = uint8(length >> 8)
	bs[3] = uint8(length)
	bs[4] = this.Control
	copy(bs[5:], this.Data)
	bs[length-1] = Sum(bs)
	return bs
}

func Decode(bs []byte) (*Frame, error) {
	if len(bs) < 6 {
		return nil, errors.New("数据长度不足")
	}

	if conv.Uint16(bs[:2]) != Prefix {
		return nil, errors.New("无效帧头")
	}

	length := conv.Uint16(bs[2:4])

	if len(bs) != int(length+6) {
		return nil, errors.New("数据长度错误")
	}

	if Sum(bs[5:length+5]) != bs[length+5] {
		return nil, errors.New("校验错误")
	}

	return &Frame{
		Control: bs[4],
		Data:    bs[5 : length+5],
	}, nil
}

func ReadFrom(r io.Reader) ([]byte, error) {

	for {

		prefix := make([]byte, 2)
		n, err := r.Read(prefix)
		if err != nil {
			return nil, err
		}
		if n != 2 {
			continue
		}

		lengthBs := make([]byte, 2)
		n, err = r.Read(lengthBs)
		if err != nil {
			return nil, err
		}
		if n != 2 {
			continue
		}
		length := conv.Uint16(lengthBs) + 2

		result := make([]byte, length+6)
		_, err = io.ReadAtLeast(r, result[4:], int(length))
		if err != nil {
			return nil, err
		}

		copy(result[0:2], prefix)
		copy(result[2:4], lengthBs)

		return result, nil

	}

}
