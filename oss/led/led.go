package led

import (
	"context"
	"machine"
	"time"
)

//// Default 默认led,部分没有,或者有多个
//func Default() *LED {
//	return New(machine.LED)
//}

func New(pin machine.Pin) *LED {
	pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	return &LED{
		Pin:   pin,
		ch:    make(chan Blink, 1),
		timer: time.NewTimer(0),
	}
}

type LED struct {
	machine.Pin                    //引脚
	ch          chan Blink         //中途插入规则
	cancel      context.CancelFunc //上下文
	timer       *time.Timer
}

// Insert 插入闪烁规则
func (this *LED) Insert(r Blink) {
	select {
	case this.ch <- r:
	default:
	}
}

// Rapid 急促闪烁一次
func (this *LED) Rapid() {
	this.Insert(BlinkRapid)
}

// Stop 停止背景led效果
func (this *LED) Stop() {
	if this.cancel != nil {
		this.cancel()
	}
}

// GoBlink 闪烁,背景led效果
func (this *LED) GoBlink(low, high float32) *LED {
	go this.Blink(low, high)
	return this
}

// Blink 闪烁,背景led效果
func (this *LED) Blink(low, high float32) {
	this.Background(Blink{
		time.Millisecond * time.Duration(low*1000),
		time.Millisecond * time.Duration(high*1000),
	})
}

// GoBackground 背景led效果
func (this *LED) GoBackground(r Blink) *LED {
	go this.Background(r)
	return this
}

// Background 背景led效果
func (this *LED) Background(r Blink) {
	this.Stop()
	var ctx context.Context
	ctx, this.cancel = context.WithCancel(context.Background())
	if len(r) == 0 {
		this.Pin.Low()
		return
	}
	if len(r) == 0 {
		r = Blink{time.Second}
	}
	for {
		select {
		case <-ctx.Done():
			return
		case insert, ok := <-this.ch:
			if !ok {
				return
			}
			this.exec(ctx, insert)
		default:
			this.exec(ctx, r)
		}
	}
}

// exec 执行led规则,半途插入效果,会立即执行,并丢弃正在执行的效果
func (this *LED) exec(ctx context.Context, r Blink) {
	for i := range r {
		this.timer.Reset(r[i])
		select {
		case <-ctx.Done():
			return
		case insert := <-this.ch:
			this.exec(ctx, insert)
			return
		case <-this.timer.C:
			if i%2 == 0 {
				this.Pin.High()
			} else {
				this.Pin.Low()
			}
		}
	}
}

// Blink 闪烁规则,效果
type Blink []time.Duration

var (
	// BlinkDefault 默认闪烁规则
	BlinkDefault = Blink{
		time.Millisecond * 500,
		time.Millisecond * 500,
	}

	// BlinkRunning 运行中闪烁规则
	BlinkRunning = Blink{
		time.Millisecond * 1500,
		time.Millisecond * 500,
	}

	// BlinkRapid 急促闪烁
	BlinkRapid = Blink{
		time.Millisecond * 200,
		time.Millisecond * 200,
	}

	// BlinkFast 快速闪烁3次
	BlinkFast = Blink{
		time.Millisecond * 100,
		time.Millisecond * 100,
		time.Millisecond * 100,
		time.Millisecond * 100,
		time.Millisecond * 100,
		time.Millisecond * 100,
	}
)

// NewRule 新建闪烁规则,1表示1秒
func NewRule(f ...float32) Blink {
	r := make(Blink, len(f))
	for i := range f {
		r[i] = time.Millisecond * time.Duration(f[i]*1000)
	}
	return r
}
