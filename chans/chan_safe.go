package chans

import (
	"github.com/injoyai/tinygo/conv"
	"github.com/injoyai/tinygo/safe"
)

func NewSafe(cap ...uint) *Safe {
	c := make(chan interface{}, conv.GetDefaultUint(0, cap...))
	return &Safe{
		C: c,
		Closer: safe.NewCloser().SetCloseFunc(func(err error) error {
			close(c)
			return nil
		}),
	}
}

type Safe struct {
	C
	*safe.Closer
}

func (this *Safe) Close() error {
	return this.Closer.Close()
}
