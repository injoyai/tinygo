package cfg

import (
	"errors"
	"github.com/injoyai/tinygo/conv"
	"github.com/injoyai/tinygo/logs"
	"io"
	"strings"
	"sync"
	"unsafe"
)

func New(r io.ReadWriter) *Cfg {
	c := &Cfg{
		r: r,
		m: make(map[string]string),
	}
	c.Extend = conv.NewExtend(c)

	if r != nil {
		buf := make([]byte, 1024)
		n, err := c.r.Read(buf)
		if err == nil {
			lines := strings.Split(string(buf[:n]), Split)
			for _, line := range lines {
				kv := strings.SplitN(line, Link, 2)
				if len(kv) == 2 {
					c.m[kv[0]] = kv[1]
				}
			}
		}
		logs.PrintErr(err)
	}

	return c
}

const (
	Link  = "="
	Split = "#"
)

type Cfg struct {
	r  io.ReadWriter
	m  map[string]string
	mu sync.RWMutex
	conv.Extend
}

func (this *Cfg) GetVar(key string) *conv.Var {
	this.mu.RLock()
	defer this.mu.RUnlock()
	return conv.New(this.m[key])
}

func (this *Cfg) Set(key string, val any) error {
	logs.Debug("SET", key, val)
	this.mu.Lock()
	this.m[key] = conv.String(val)
	this.mu.Unlock()
	return this.Save()
}

func (this *Cfg) Save() error {
	if this.r == nil {
		return errors.New("无效IO")
	}
	logs.Debug("Write", string(this.Bytes()))
	_, err := this.r.Write(this.Bytes())
	return err
}

func (this *Cfg) Bytes() []byte {
	s := ""
	this.mu.Lock()
	for k, v := range this.m {
		s += k + Link + v + Split
	}
	this.mu.Unlock()
	return *(*[]byte)(unsafe.Pointer(&s))
}
