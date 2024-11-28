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

func New(pin machine.Pin, background ...Rule) *LED {
	bg := RuleNull
	if len(background) > 0 {
		bg = background[0]
	}
	pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	return &LED{
		Pin:        pin,
		background: bg,
	}
}

type LED struct {
	machine.Pin                    //引脚
	ch          chan Rule          //
	cancel      context.CancelFunc //上下文
	background  Rule
}

// Blink 闪烁
func (this *LED) Blink(low, high float32) {
	this.Exec(Rule{
		time.Millisecond * time.Duration(low*1000),
		time.Millisecond * time.Duration(high*1000),
	})
}

func (this *LED) Once() {

}

// Exec 执行闪烁规则
func (this *LED) Exec(r Rule) {
	this.ExecWithContext(context.Background(), r)
}

// ExecWithContext 执行闪烁规则
func (this *LED) ExecWithContext(ctx context.Context, r Rule) {
	if this.cancel != nil {
		this.cancel()
	}
	ctx, this.cancel = context.WithCancel(ctx)
	if len(r) == 0 {
		this.Pin.Low()
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		default:
			for i := range r {
				select {
				case <-ctx.Done():
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
	}
}

// Rule 闪烁规则
type Rule []time.Duration

var (
	RuleNull = []time.Duration{time.Second}

	// RuleDefault 默认闪烁规则
	RuleDefault = []time.Duration{
		time.Millisecond * 500,
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
