package easy

import (
	"bufio"
	"errors"
	"github.com/injoyai/tinygo/maps"
	"io"
)

func NewServer() *Server {
	return &Server{
		m: maps.NewSafe(),
	}
}

type Handler struct {
	Get func() any
	Set func(value any) error
}

type Server struct {
	m *maps.Safe
}

func (this *Server) Register(k string, h Handler) {
	this.m.Set(k, h)
}

func (this *Server) Bridge(r io.ReadWriter) error {
	buf := bufio.NewReaderSize(r, 128)
	for {
		f, err := ReadFrom(buf)
		if err != nil {
			return err
		}
		switch f.Type {
		case GET:
			h, ok := this.m.Get(f.Key)
			if !ok {
				r.Write(f.Resp(errors.New("not found")).Bytes())
			} else {
				val := h.(Handler).Get()
				r.Write(f.Resp(val).Bytes())
			}

		case SET:
			h, ok := this.m.Get(f.Key)
			if !ok {
				r.Write(f.Resp(errors.New("not found")).Bytes())
			} else {
				err = h.(Handler).Set(f.Value)
				r.Write(f.Resp(err).Bytes())
			}

		}
	}
}
