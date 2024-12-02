package serial

import (
	"github.com/injoyai/tinygo/logs"
	"io"
	"time"
)

//type (
//	Option = client.Option
//	Server = client.Client
//)

func NewServer(p *Serial, bufSize uint16, op ...Option) *Server {
	s := &Server{
		p:   p,
		buf: make([]byte, bufSize),
	}
	for _, v := range op {
		v(s)
	}
	return s
}

type Option func(s *Server)

type Server struct {
	p             *Serial
	buf           []byte
	OnDealMessage func(bs []byte)
}

func (this *Server) GoTimerWriter(t time.Duration, f func(w io.Writer)) {
	go func() {
		for {
			<-time.After(t)
			f(this.p)
		}
	}()
}

func (this *Server) Run() {

	for {
		n, err := this.p.Read(this.buf)
		if err != nil {
			logs.Err(err)
			continue
		}
		if n == 0 {
			<-time.After(time.Millisecond * 10)
			continue
		}
		if n > 0 && this.OnDealMessage != nil {
			this.OnDealMessage(this.buf[:n])
		}
	}
}
