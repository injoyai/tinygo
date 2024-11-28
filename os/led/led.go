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
		Pin: pin,
		ch:  make(chan Rule, 1),
	}
}

type LED struct {
	machine.Pin                    //引脚
	ch          chan Rule          //中途插入规则
	cancel      context.CancelFunc //上下文
}

// Insert 插入闪烁规则
func (this *LED) Insert(r Rule) {
	select {
	case this.ch <- r:
	default:
	}
}

// Rapid 急促闪烁一次
func (this *LED) Rapid() {
	this.Insert(RuleRapid)
}

// Blink 闪烁,背景led效果
func (this *LED) Blink(low, high float32) {
	this.Background(Rule{
		time.Millisecond * time.Duration(low*1000),
		time.Millisecond * time.Duration(high*1000),
	})
}

func (this *LED) Stop() {
	if this.cancel != nil {
		this.cancel()
	}
}

// Background 背景led效果
func (this *LED) Background(r Rule) {
	this.Stop()
	var ctx context.Context
	ctx, this.cancel = context.WithCancel(context.Background())
	if len(r) == 0 {
		this.Pin.Low()
		return
	}
	if len(r) == 0 {
		r = Rule{time.Second}
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
func (this *LED) exec(ctx context.Context, r Rule) {
	for i := range r {
		select {
		case <-ctx.Done():
			return
		case insert := <-this.ch:
			this.exec(ctx, insert)
			return
		case <-time.After(r[i]):
			if i%2 == 0 {
				this.Pin.High()
			} else {
				this.Pin.Low()
			}
		}
	}
}

// Rule 闪烁规则,效果
type Rule []time.Duration

var (
	// RuleDefault 默认闪烁规则
	RuleDefault = Rule{
		time.Millisecond * 1500,
		time.Millisecond * 500,
	}

	// RuleRapid 急促闪烁
	RuleRapid = Rule{
		time.Millisecond * 200,
		time.Millisecond * 200,
	}
)

// NewRule 新建闪烁规则,1表示1秒
func NewRule(f ...float32) Rule {
	r := Rule{}
	for _, v := range f {
		r = append(r, time.Millisecond*time.Duration(v*1000))
	}
	return r
}
