package client

import (
	"bufio"
	"context"
	"errors"
	"github.com/injoyai/tinygo/bytes"
	"github.com/injoyai/tinygo/ios"
	"github.com/injoyai/tinygo/ios/module/common"
	"github.com/injoyai/tinygo/maps"
	"github.com/injoyai/tinygo/safe"
	"io"
	"time"
)

type (
	Option  func(c *Client)
	Message = bytes.Entity
)

func Run(f ios.DialFunc, op ...Option) error {
	c, err := Dial(f, op...)
	if err != nil {
		return err
	}
	return c.Run()
}

func Redial(f ios.DialFunc, op ...Option) *Client {
	return RedialWithContext(context.Background(), f, op...)
}

func RedialWithContext(ctx context.Context, dial ios.DialFunc, op ...Option) *Client {
	c := NewWithContext(ctx)
	c.SetRedial()
	_ = c.MustDial(dial, op...)
	return c
}

func Dial(f ios.DialFunc, op ...Option) (*Client, error) {
	return DialWithContext(context.Background(), f, op...)
}

func DialWithContext(ctx context.Context, dial ios.DialFunc, op ...Option) (*Client, error) {
	c := NewWithContext(ctx)
	err := c.Dial(dial, op...)
	return c, err
}

func New() *Client {
	return NewWithContext(context.Background())
}

func NewWithContext(ctx context.Context) *Client {
	c := &Client{
		key:        "",
		Reader:     nil,
		MoreWriter: nil,
		Logger:     common.NewLogger(),
		Info: Info{
			CreateTime: time.Now(),
			DialTime:   time.Now(),
		},
		Event: &Event{
			OnDealErr: func(c *Client, err error) error { return common.DealErr(err) },
		},
		Closer:     safe.NewCloser(),
		Runner:     safe.NewRunnerWithContext(ctx, nil),
		Tag:        maps.NewSafe(),
		timeout:    safe.NewRunnerWithContext(ctx, nil),
		Ctx:        ctx,
		redialSign: make(chan struct{}),
		dialedSign: make(chan struct{}),
		dial:       nil,
		options:    nil,
	}
	c.Closer.CloseWithErr(errors.New("等待连接"))
	c.Runner.SetFunc(c.run)
	return c
}

/*
Client
客户端的指针地址是唯一标识,key是表面的唯一标识,需要用户自己维护
*/
type Client struct {
	ios.Reader     //IO实例 目前支持ios.AReader,ios.MReader,io.Reader
	ios.AllReader  //实现多种读取方式
	ios.MoreWriter //多个方式写入

	Info                         //基本信息
	*Event                       //事件
	*safe.Closer                 //关闭
	*safe.Runner                 //运行,全局的生命周期,包括重试
	Logger       common.Logger   //日志
	Tag          *maps.Safe      //标签,用于记录连接的一些信息
	Ctx          context.Context //上下文
	timeout      *safe.Runner    //超时机制

	key        string        //自定义标识
	redial     bool          //是否自动重连
	redialSign chan struct{} //重连信号,未设置自动重连也可以手动重连
	dialedSign chan struct{} //已连接信号,连接的时候会释放信号,关闭的时候会重新阻塞,注意不是一个
	dial       ios.DialFunc  //连接函数
	options    []Option      //选项
}

// SetBuffer 仅对io.Reader有效
func (this *Client) SetBuffer(size int) *Client {
	switch v := this.Reader.(type) {
	case ios.MReader:
	case ios.AReader:
	case io.Reader:
		this.Reader = bufio.NewReaderSize(v, size)
	}
	return this
}

func (this *Client) SetReadWriteCloser(k string, r ios.ReadWriteCloser, op ...Option) {
	this.key = k
	this.Reader = r
	this.SetBuffer(4096) //设置缓存区4KB,能大幅度提升性能
	//需要先初始化，方便OnConnect的数据读取,run的时候还会声明一次最新的读取函数
	this.AllReader = ios.NewAllReader(this.Reader, this.Event.OnReadFrom)
	this.MoreWriter = ios.NewMoreWriter(r)
	this.Info.DialTime = time.Now()
	this.options = op
	//Runner现在作为全局的生命周期,Closer控制单次的生命周期
	//this.Runner = safe.NewRunnerWithContext(this.Ctx, this.run)
	this.Closer = safe.NewCloser().SetCloseFunc(func(err error) error {

		//关闭真实实例
		if er := r.Close(); er != nil {
			return er
		}
		//关闭超时机制
		this.timeout.Stop()

		//关闭/断开连接事件
		this.Logger.Errorf("[%s] 断开连接: %s\n", this.GetKey(), err.Error())
		if this.Event.OnDisconnect != nil {
			this.Event.OnDisconnect(this, err)
		}
		return nil
	})
	//写入事件
	this.MoreWriter.(*ios.MoreWrite).Option = []ios.WriteOption{
		func(p []byte) (_ []byte, err error) {
			this.Logger.Writeln("["+this.GetKey()+"] ", p)
			if this.Event.OnWriteWith != nil {
				p, err = this.Event.OnWriteWith(p)
			}
			this.Info.WriteTime = time.Now()
			this.Info.WriteCount++
			this.Info.WriteBytes += len(p)
			return p, err
		},
	}

	//执行选项
	this.SetOption(op...)

}

func (this *Client) MustDial(dial ios.DialFunc, op ...Option) error {
	return this._dial(true, dial, op...)
}

func (this *Client) Dial(dial ios.DialFunc, op ...Option) error {
	return this._dial(false, dial, op...)
}

func (this *Client) _dial(must bool, dial ios.DialFunc, op ...Option) (err error) {
	this.dial = dial
	r, k, err := this.doDial(must)
	if err != nil {
		this.Closer = safe.NewCloser()
		this.Closer.CloseWithErr(err)
		return err
	}
	//增加连接成功的信号,方便一些逻辑判断
	//当连接成功的时候会发送信号通道(防止代码错误最多10个),则监听能收到连接成功的信号
	//当连接关闭的时候,需要重新声明信号,方便下次阻塞
	for i := 0; i < 10; i++ {
		select {
		case this.dialedSign <- struct{}{}:
		default:
			break
		}
	}
	this.SetReadWriteCloser(k, r, op...)
	//连接事件
	this.Logger.Infof("[%s] 连接服务成功...\n", this.GetKey())
	if this.Event.OnConnected != nil {
		if err := this.Event.OnConnected(this); err != nil {
			this.CloseWithErr(err)
			return err
		}
	}
	return nil
}

func (this *Client) doDial(must bool) (ios.ReadWriteCloser, string, error) {
	if this.dial == nil {
		return nil, "", errors.New("handler is nil")
	}
	if !must {
		return this.dial(this.Ctx)
	}
	this.Logger.Infof("等待连接服务...\n")
	if this.Event != nil && this.Event.OnReconnect != nil {
		return this.Event.OnReconnect(this, this.dial)
	}
	//防止用户设置错了重试,再外层在加上一层退避重试,是否需要? 可能想重试10次就不重试就无法实现了
	f := ReconnectWithRetreat(time.Second*2, time.Second*32, 2)
	return f(this, this.dial)
}

// SetReadTimeout 设置读取超时,即距离上次读取数据时间超过该设置值,则会关闭连接,0表示不超时
func (this *Client) SetReadTimeout(timeout time.Duration) *Client {
	this.timeout.SetFunc(func(ctx context.Context) error {
		if timeout <= 0 {
			return nil
		}
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		for {
			timer.Reset(timeout)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				this.CloseWithErr(ios.ErrReadTimeout)
				return ios.ErrReadTimeout
			}
		}
	})

	//不用判断客户端是否已经运行,可能还没开始执行
	//if this.timeout.Running() {
	this.timeout.Restart()
	//}

	return this
}

func (this *Client) SetOption(op ...Option) *Client {
	for _, fn := range op {
		fn(this)
	}
	return this
}

func (this *Client) GetKey() string {
	return this.key
}

func (this *Client) SetKey(key string) *Client {
	oldKey := this.key
	this.key = key
	if this.key != oldKey {
		this.Logger.Infof("[%s] 修改标识为 [%s]\n", oldKey, this.key)
		if this.Event.OnKeyChange != nil {
			this.Event.OnKeyChange(this, oldKey)
		}
	}
	return this
}

func (this *Client) Timer(t time.Duration, f func(c *Client)) {
	tick := time.NewTicker(t)
	defer tick.Stop()
	for {
		select {
		case <-this.Closer.Done():
			return
		case _, ok := <-tick.C:
			if ok {
				f(this)
			}
		}
	}
}

// GoTimerWriter 定时写入,容易忘记使用协程,然后阻塞,索性直接用协程
func (this *Client) GoTimerWriter(t time.Duration, f func(w ios.MoreWriter) error) {
	go this.Timer(t, func(c *Client) {
		err := f(c)
		if this.Event != nil && this.Event.OnDealErr != nil {
			err = this.Event.OnDealErr(c, err)
		}
		c.CloseWithErr(err)
	})
}

// CloseAll 关闭连接,并不再重试
func (this *Client) CloseAll() error {
	this.SetRedial(false)
	return this.Closer.Close()
}

// SetRedial 设置自动重连,当连接断开时,
// 会进行自动重连,退避重试,直到成功,除非上下文关闭
func (this *Client) SetRedial(b ...bool) *Client {
	this.redial = len(b) == 0 || b[0]
	return this
}

// Redial 断开重连,是否有必要? 因为可以用其他方式实现
func (this *Client) Redial() {
	this.redialSign <- struct{}{}
}

// Dialed 连接成功的信号,每次连接成功都会释放信号(最多10个)
// 需要判断下closer是否关闭,才能保证逻辑更有效
func (this *Client) Dialed() <-chan struct{} {
	return this.dialedSign
}

// Done 这个是客户端生命周期结束的关闭信号
func (this *Client) Done() <-chan struct{} {
	return this.Runner.Done()
}

// run 运行读取数据操作,如果设置了重试,则这个run结束后立马执行run,递归下去,是否会有资源未释放?
func (this *Client) run(ctx context.Context) (err error) {

	//校验事件函数
	if this.Event == nil {
		this.Event = &Event{}
	}

	//运行的时候,重新加载下OnReadFrom
	this.AllReader = ios.NewAllReader(this.Reader, this.Event.OnReadFrom)

	//超时机制
	this.timeout.Start()

	for {
		select {

		case <-ctx.Done():
			//这个ctx不是Client的ctx,而是Runner的ctx属于Client的ctx的子级
			return this.Closer.Err()

		case <-this.Closer.Done():
			//一个连接的生命周期结束
			if this.redial {
				//设置了重连,并且已经运行,其他都关闭
				//这里连接的错误只会出现在上下文关闭的情况
				if err := this.MustDial(this.dial, this.options...); err != nil {
					return err
				}
				return this.run(ctx)
			}
			return this.Closer.Err()

		case <-this.redialSign:

			//先关闭老连接
			this.CloseWithErr(errors.New("手动重连"))
			//尝试建立连接,不需要重试,连接失败后会进行下一个循环
			//下个循环会走正常的断开是否重连逻辑,设置重连会一直重试,否则退出执行
			this.Dial(this.dial, this.options...)

		default:

			//读取数据,目前支持3种类型,Reader, AReader, MReader
			//如果是AReader,MReader,说明是分包分好的数据,则直接读取即可
			//如果是Reader,则数据还处于粘包状态,需要调用时间OnReadBuffer,来进行读取
			ack, err := this.ReadAck()
			if err != nil {
				if this.Event.OnDealErr != nil {
					err = this.Event.OnDealErr(this, err)
				}
				this.CloseWithErr(err)
				//交给closer进行处理接下来的逻辑,固这里不使用return
				//例如重新连接等操作,这样只用写一个地方,简化代码
				continue
			}
			this.Info.ReadTime = time.Now()
			this.Info.ReadCount++
			this.Info.ReadBytes += len(ack.Payload())

			//处理数据,使用事件OnDealMessage处理数据,
			//如果未实现,则不处理数据,则不会确认消息
			this.Logger.Readln("["+this.GetKey()+"] ", ack.Payload())
			if this.Event.OnDealMessage != nil {
				this.Event.OnDealMessage(this, ack)
			}

		}
	}

}
