package easy

import (
	"github.com/injoyai/tinygo/conv"
	"io"
	"strings"
	"unsafe"
)

/*

1.
GET version
GET ok v1.0.1
GET err 错误信息

SET version v1.0.2
SET version ok v1.0.2
SET version err 错误信息






*/

const (
	Spilt = " "
	GET   = "GET"
	SET   = "SET"
)

type Request struct {
	Type  string //GET SET
	Key   string
	Value any
}

func (this *Request) Bytes() []byte {

	switch this.Type {
	case GET:
		s := GET + Spilt + this.Key
		return *(*[]byte)(unsafe.Pointer(&s))

	case SET:
		s := SET + Spilt + this.Key + Spilt + conv.String(this.Value)
		return *(*[]byte)(unsafe.Pointer(&s))

	default:
		s := GET + Spilt + this.Key
		return *(*[]byte)(unsafe.Pointer(&s))

	}
}

func ReadFrom(r io.Reader) (*Request, error) {
	for {
		bs, err := readFrom(r)
		if err != nil {
			return nil, err
		}
		s := *(*string)(unsafe.Pointer(&bs))
		ls := strings.SplitN(s, Spilt, 3)
		if len(ls) < 2 {
			continue
		}
		req := &Request{
			Type: ls[0],
			Key:  ls[1],
		}
		if len(ls) >= 3 {
			req.Value = ls[2]
		}
		return req, nil
	}
}

func readFrom(r io.Reader) ([]byte, error) {

	buf := make([]byte, 3)
	for {
		_, err := io.ReadAtLeast(r, buf, len(buf))
		if err != nil {
			return nil, err
		}

		switch string(buf) {
		case GET, SET:

			b := make([]byte, 1)
			for {
				_, err = io.ReadAtLeast(r, b, len(b))
				if err != nil {
					return nil, err
				}
				if b[0] == '\n' {
					if buf[len(buf)-1] == '\r' {
						return buf[:len(buf)-1], nil
					}
					return buf, nil
				}
				buf = append(buf, b[0])
			}
		}

	}
}

func (this *Request) Resp(v any) *Response {
	return &Response{
		Type:  this.Type,
		Key:   this.Key,
		Value: v,
	}
}

type Response struct {
	Type  string
	Key   string
	Value any
}

func (this *Response) Bytes() []byte {
	s := this.Type + Spilt + this.Key + Spilt
	switch val := this.Value.(type) {
	case error:
		s += "err" + Spilt + val.Error()
	default:
		s += "ok" + Spilt + conv.String(this.Value)
	}
	s += "\n"
	return *(*[]byte)(unsafe.Pointer(&s))
}
