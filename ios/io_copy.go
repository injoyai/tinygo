package ios

import (
	"io"
)

// Bridge 桥接,桥接两个ReadWriter
// 例如,桥接串口(客户端)和网口(tcp客户端),可以实现通过串口上网
func Bridge(i1, i2 io.ReadWriter) error {
	return Swap(i1, i2)
}

func Swap(r1, r2 io.ReadWriter) error {
	go Copy(r1, r2)
	_, err := Copy(r2, r1)
	return err
}

func Copy(w io.Writer, r io.Reader) (int64, error) {
	return io.Copy(w, r)
}

func CopyBuffer(w io.Writer, r io.Reader, buf []byte) (int64, error) {
	return io.CopyBuffer(w, r, buf)
}

func CopyWith(w io.Writer, r io.Reader, f func(p []byte) ([]byte, error)) (int64, error) {
	return CopyBufferWith(w, r, nil, f)
}

// CopyBufferWith 复制数据,每次固定大小,并提供函数监听
// 如何使用接口约束 [T Reader | MReader | AReader]
func CopyBufferWith(w io.Writer, r Reader, buf []byte, f func(p []byte) ([]byte, error)) (int64, error) {

	read := NewReadFrom(buf)

	for co, n := int64(0), 0; ; co += int64(n) {
		bs, err := read(r)
		if err != nil {
			if err == io.EOF {
				return co, nil
			}
			return 0, err
		}
		if f != nil {
			bs, err = f(bs)
			if err != nil {
				return 0, err
			}
		}
		n, err = w.Write(bs)
		if err != nil {
			return 0, err
		}
	}

}

func ReadBuffer(r Reader, buf []byte) (Acker, error) {
	readFunc := NewReadFrom(buf)
	bs, err := readFunc(r)
	return Ack(bs), err
}
