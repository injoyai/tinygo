package easy

import (
	"bufio"
	"errors"
	"github.com/injoyai/tinygo/cfg"
	"io"
	"strings"
	"sync"
)

func NewServer() *Server {
	return &Server{
		m: make(map[string]Handler),
	}
}

type Handler struct {
	Get func() any
	Set func(value any) error
}

type Server struct {
	m  map[string]Handler
	mu sync.RWMutex
}

// Register 注册
// 使用runtime.ReadMemStats会卡死
func (this *Server) Register(k string, h Handler) {
	this.mu.Lock()
	this.m[strings.ToUpper(k)] = h
	this.mu.Unlock()
}

func (this *Server) RegisterCfg(k string) {
	this.Register(k, Handler{
		Get: func() any { return cfg.Get("test") },
		Set: func(value any) error { return cfg.Set("test", value) },
	})
}

func (this *Server) RegisterGet(k string, f func() any) {
	this.Register(k, Handler{Get: f})
}

func (this *Server) Bridge(r io.ReadWriter) error {
	buf := bufio.NewReaderSize(r, 128)
	for {
		f, err := ReadFrom(buf)
		if err != nil {
			return err
		}
		f.Key = strings.ToUpper(f.Key)
		switch f.Type {
		case GET:
			this.mu.RLock()
			h, ok := this.m[f.Key]
			this.mu.RUnlock()
			if !ok {
				r.Write(f.Resp(errors.New("not found")).Bytes())
			} else {
				val := h.Get()
				r.Write(f.Resp(val).Bytes())
			}

		case SET:
			this.mu.RLock()
			h, ok := this.m[f.Key]
			this.mu.RUnlock()
			if !ok {
				r.Write(f.Resp(errors.New("not found")).Bytes())
			} else {
				err = h.Set(f.Value)
				r.Write(f.Resp(err).Bytes())
			}

		}
	}
}
