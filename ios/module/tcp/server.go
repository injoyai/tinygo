package tcp

import (
	"fmt"
	"github.com/injoyai/tinygo/ios"
	"net"
)

var _ ios.Listener = (*Server)(nil)

func NewListen(port int) func() (ios.Listener, error) {
	return func() (ios.Listener, error) {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return nil, err
		}
		return &Server{
			Listener: listener,
		}, nil
	}
}

type Server struct {
	net.Listener
}

func (this *Server) Close() error {
	return this.Listener.Close()
}

func (this *Server) Accept() (ios.ReadWriteCloser, string, error) {
	c, err := this.Listener.Accept()
	if err != nil {
		return nil, "", err
	}
	return c, c.RemoteAddr().String(), nil
}

func (this *Server) Addr() string {
	return this.Listener.Addr().String()
}
