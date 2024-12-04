package ec800e

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/injoyai/tinygo/conv"
	"github.com/injoyai/tinygo/logs"
	"github.com/injoyai/tinygo/maps/wait"
	"io"
	"strings"
	"sync"
	"time"
)

// New 4G模块断开网络很慢,可能要好几分钟,初始化的时候无法等待这么久
// 实际使用可以通过上下电的方式重置网络,以达到和缓存配置同步的目的
// todo 待去文档查找一键重置命令
func New(r io.ReadWriteCloser, timeout time.Duration) *AT {
	at := &AT{
		r:    r,
		wait: wait.New(timeout),
		ignoreErr: map[string]struct{}{
			"serial: timeout": {},
		},
		tcp:  [12]*TCP{},
		MQTT: [6]*MQTT{},
	}

	for i := int8(0); i < 6; i++ {
		at.MQTT[i] = newMQTT(at, i)
	}

	go at.run()
	//go func() {
	//	//初始化操作
	//	//1. 关闭所有MQTT客户端
	//	for i := int8(0); i < 6; i++ {
	//		at.write(at.encode("+QMTDISC=" + conv.String(i)))
	//		at.write(at.encode("+QMTCLOSE=" + conv.String(i)))
	//	}
	//	//设置socket关闭时,自动关闭客户端
	//	at.encodeWrite(`+QICFG="passiveclosed",1`)
	//	//2. 关闭所有TCP客户端
	//	for i := int8(0); i <= 11; i++ {
	//		at.write(at.encode("+QICLOSE=" + conv.String(i)))
	//	}
	//}()
	return at
}

/*
AT 命令
参考
https://blog.csdn.net/weixin_71997855/article/details/139340415
*/
type AT struct {
	r         io.ReadWriteCloser  //IO,例如串口
	wait      *wait.Entity        //异步等待
	ignoreErr map[string]struct{} //忽略错误,继续执行,例串口
	mu        sync.Mutex          //锁,防止并发
	tcp       [12]*TCP            //tcp客户端
	MQTT      [6]*MQTT            //mqtt客户端信息
	ICCID     string
	IMEI      string
	IPv4      string
}

func (this *AT) GetMQTT(clientIndex int8) *MQTT {
	return this.MQTT[clientIndex]
}

func (this *AT) SetIO(r io.ReadWriteCloser) *AT {
	this.r = r
	return this
}

/*
Run
需要打开回显使用,回显作为消息的唯一标识
同步的会有这个消息ID,异步的为空,
异步的通过消息内容来区分,格式为 +XXX: xxx
数据分包是根据2个\r\n

响应情况1:
----------------------------------------------
ATI
Quectel
EC801E
Revision: EC801ECNCGR03A03M02

# OK

----------------------------------------------

响应情况2:
----------------------------------------------
AT+QPING=1,"39.107.120.124"
OK

----------------------------------------------
*/
func (this *AT) run() error {
	r := bufio.NewReaderSize(this.r, 1601)
	buf := make([]byte, 0)
	/*
		这个数据读取比较麻烦
		1. 根据结尾来读取数据,当结尾是以下数据时,算一个报文
			正常响应:"\r\nOK\r\n"
			异常响应:"\r\nERROR\r\n"
			启动成功: "\r\nRDY\r\n"
			TCP透传异常: "\r\nNO CARRIER\r\n"
			等待输入: "\r\n>"

		2. 想MQTT订阅消息的格式如下
			\r\nAT+QMTRECV: 0,0,"test","2024-08-21 14:18:48.034039 +0800 CST m=+900.375396501"\r\n

			当开头和结尾都是"\r\n"时,算一个报文




	*/
	for {
		bs, err := this.readBuffer(r, buf)
		if err != nil {
			if _, ok := this.ignoreErr[err.Error()]; ok {
				continue
			}
			logs.Err(err)
			return err
		}
		this.deal(bs)
	}
}

func (this *AT) readBuffer(r *bufio.Reader, buf []byte) ([]byte, error) {
	buf = buf[:0]
	for {
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		buf = append(buf, b)

		if len(buf) >= 3 {

			switch {
			case bytes.HasSuffix(buf, []byte("\r\nOK\r\n")):
				//正常响应数据
				return buf[:len(buf)-2], nil

			case bytes.HasSuffix(buf, []byte("\r\nSEND OK\r\n")):
				//TCP发送生成响应数据
				return buf[:len(buf)-2], nil

			case bytes.HasSuffix(buf, []byte("\r\nERROR\r\n")):
				//异常响应数据
				return buf[:len(buf)-2], nil

			case bytes.HasSuffix(buf, []byte("\r\nRDY\r\n")):
				//启动成功
				return buf[:len(buf)-2], nil

			case bytes.HasSuffix(buf, []byte("\r\n>")):
				//等待输入
				return buf[:], nil

			case bytes.HasSuffix(buf, []byte("\r\nNO CARRIER\r\n")):
				//透传异常
				return buf[:len(buf)-2], nil

			case bytes.HasPrefix(buf, []byte("\r\n+QIURC: \"recv\",")) && bytes.HasSuffix(buf, []byte("\r\n")):
				//收到tcp数据
				//1. 获取数据长度

				if i := bytes.IndexByte(buf[17:], ','); i > 0 {
					if i2 := bytes.Index(buf[17+i+1:], []byte("\r\n")); i2 > 0 {
						x := buf[17+i+1 : 17+i+1+i2]
						length := conv.Int(string(x))
						if len(buf[17+i+1+i2:]) == length+4 {
							return append(buf[:17+i+1], buf[17+i+1+i2+2:]...), nil
						}
					}
				}

			case bytes.HasPrefix(buf, []byte("\r\n")) && bytes.HasSuffix(buf, []byte("\r\n")):
				//其他数据,例回调 头尾都是"\r\n",不能讲正常响应数据拆分处理
				return buf[:len(buf)-2], nil

			default:

			}

		}

	}
}

func (this *AT) deal(bs []byte) {
	if len(bs) == 0 {
		return
	}
	logs.Read(string(bs))
	ls := bytes.Split(bs, []byte("\r\n"))

	waitKey := ""
	for i, v := range ls {

		s := string(v)
		switch {

		case len(s) == 0:
			continue

		case i == 0: //&& strings.HasPrefix(s, "AT"):
			//第一行,如果设置了回显(默认),则数据内容是之前发送的数据
			//否则为空(未设置回显或者回调),自带一个\r,则补充个\n
			if len(v) > 0 && v[len(v)-1] == '\r' {
				v = append(v, '\n')
			}
			waitKey, _ = strings.CutPrefix(string(v), " ")

		case s == "OK":
			//下发命令的响应成功
			this.wait.Done(waitKey, nil)

		case s == "SEND OK":
			//TCP发送数据响应成功
			this.wait.Done(waitKey, nil)

		case s == ">":
			//数据输入
			this.wait.Done(waitKey, nil)
			continue

		case s == "ERROR":
			//下发命令响应失败
			this.wait.Done(waitKey, errors.New("响应错误: "+s))

		case s == "RDY":
			//模块发生重启的情况
			//重启的回调,全部重新加载?
			//关闭全部连接,实际socket已经断开,重试机制交给业务逻辑
			for _, c := range this.MQTT {
				if c != nil && c.Closer != nil {
					c.Close()
				}
			}
			for _, c := range this.tcp {
				if c != nil && c.Closer != nil {
					c.Close()
				}
			}
			//设置socket关闭时,自动关闭客户端
			this.encodeWrite(`+QICFG="passiveclosed",1`)

		case s == "NO CARRIER":
		//这是啥错? 透传模式,服务端关闭触发的

		case strings.HasPrefix(s, "+ICCID: "):
			//ICCID数据
			this.wait.Done("+ICCID", s[8:])

		case strings.HasPrefix(s, "+CGPADDR: "):
			//ip信息
			this.wait.Done("+CGPADDR", s[10:])

		case len(s) == 15 && strings.HasPrefix(s, "86"):
			//IMEI数据
			this.wait.Done("+CGSN", s)

		case strings.HasPrefix(s, "+CSQ: "):
			//信号数据
			this.wait.Done("+CSQ", s[6:])

		case strings.HasPrefix(s, "+QLTS: "):
			//网络时间 +QLTS: "2019/01/13,03:40:48+32,0"
			if len(s[7:]) >= 20 {
				logs.Debug(s[8:27])
				this.wait.Done("+QLTS", s[8:27])
			}

		case strings.HasPrefix(s, "+QPING: "):
			//ping响应
			/*

				+QPING: 0,"39.107.120.124",64,223,255

				+QPING: 0,"39.107.120.124",64,59,255

				+QPING: 0,"39.107.120.124",64,52,255

				+QPING: 0,"39.107.120.124",64,61,255

				+QPING: 0,4,4,0,52,223,78

			*/

		case strings.HasPrefix(s, "+QIOPEN: "):
			if ls := strings.SplitN(s, ",", 2); len(ls) == 2 {
				this.wait.Done(ls[0], ls[1])
			}

		case strings.HasPrefix(s, "+QIURC: "):
			//透传网络状态响应
			// TCP 断开 +QIURC: "closed",6
			if ls := strings.SplitN(s[8:], ",", 2); len(ls) == 2 {
				switch ls[0] {
				case `"closed"`:
					//tcp断开连接
					connectID := conv.Int8(ls[1])
					if connectID < int8(len(this.tcp)) {
						if t := this.tcp[connectID]; t != nil {
							t.Close()
						}
					}
				case `"recv"`:
					//收到tcp数据
					if ll := strings.SplitN(ls[1], ",", 2); len(ll) == 2 {
						connectID := conv.Int8(ll[0])
						if connectID < int8(len(this.tcp)) {
							if t := this.tcp[connectID]; t != nil {
								t.ch.Write([]byte(ll[1]))
							}
						}
					}
				}

			}

		//=================================MQTT================================

		case strings.HasPrefix(s, "+QMTSTAT: "):
			//MQTT状态响应,错误信息
			if ls := strings.SplitN(string(v[10:]), ",", 2); len(ls) == 2 {
				mqttIndex := conv.Int8(ls[0])
				err := MQTTURCErrCode(ls[1]).Err()
				logs.Errf("[MQTT:%d] 连接断开: %v\n", mqttIndex, err)
				c := this.MQTT[mqttIndex]
				if c != nil && c.Closer != nil {
					//尝试进行
					c.Close()
					//go c.setClosed()
				}
			}

		case strings.HasPrefix(s, "+QMTRECV: "):
			//接收的MQTT订阅消息,这个消息可能和响应一起回来
			//<client_idx>,<msgid>,<topic>[,<payload_len>],<payload>
			if ls := strings.SplitN(s[10:], ",", 4); len(ls) == 4 {
				mqttIndex := conv.Int8(ls[0])
				c := this.MQTT[mqttIndex]
				if c != nil {
					//去除双引号
					if len(ls[2]) >= 2 {
						ls[2] = ls[2][1 : len(ls[2])-1]
					}
					if len(ls[3]) >= 2 {
						ls[3] = ls[3][1 : len(ls[3])-1]
					}
					c.setSubscript(ls[2], []byte(ls[3]))
				}
			}

		case strings.HasPrefix(s, "+QMTOPEN: "):
			// 打开MQTT网络连接异步回调
			// clientIndex,result
			if ls := strings.SplitN(s, ",", 2); len(ls) == 2 {
				this.wait.Done(ls[0], ls[1])
			}

		case strings.HasPrefix(s, "+QMTCONN: "):
			//MQTT连接状态 +QMTCONN: <client_idx>,<result>[,<ret_code>]
			if ls := strings.SplitN(s, ",", 3); len(ls) == 3 {
				this.wait.Done(ls[0], ls[2])
			}

		case strings.HasPrefix(s, "+QMTSUB: "):
			//订阅结果 +QMTSUB: <client_idx>,<msgid>,<result>[,<value>]
			if ls := strings.SplitN(s, ",", 4); len(ls) == 4 {
				this.wait.Done(ls[0], ls[2])
			}

		case strings.HasPrefix(s, "+QMTPUBEX: "):
			//发布消息响应 +QMTPUBEX: <client_idx>,<msgid>,<result>[,<value>]
			if ls := strings.SplitN(s, ",", 3); len(ls) == 3 {
				//result:
				//0 数据包发送成功且接收到服务器的 ACK（当<qos>=0 时发布了数据，则无需 ACK）
				//1 数据包重传
				if ls[2] == "0" {
					this.wait.Done(ls[0], ls[2])
				}
			}

		case strings.HasPrefix(s, "+QMTDISC: "):
		//断开MQTT会话层响应
		//主动关闭响应,根据"+QMTSTAT"状态来

		case strings.HasPrefix(s, "+QMTCLOSE: "):
			//关闭MQTT响应
			if ls := strings.SplitN(s, ",", 2); len(ls) == 2 {
				this.wait.Done(ls[0], ls[1])
			}

		}

	}
}

func (this *AT) encode(command string) []byte {
	s := "AT" + command + "\r\n"
	return []byte(s)
}

func (this *AT) write(p []byte) (n int, err error) {
	return this.r.Write(p)
}

func (this *AT) encodeWrite(command string) error {
	p := this.encode(command)
	logs.Write(string(p))
	_, err := this.r.Write(p)
	return err
}

func (this *AT) send2(command string) (string, error) {
	p := this.encode(command)
	logs.Write(string(p))
	_, err := this.r.Write(p)
	if err != nil {
		return "", err
	}
	result, err := this.wait.Wait(string(p))
	if err != nil {
		return "", err
	}
	return conv.String(result), nil
}

func (this *AT) send(key string, p []byte) (string, error) {
	logs.Write(string(p))
	_, err := this.r.Write(p)
	if err != nil {
		return "", err
	}
	result, err := this.wait.Wait(key)
	if err != nil {
		return "", err
	}
	return conv.String(result), nil
}

// Version 4g模块版本
func (this *AT) Version() (string, error) {
	return this.send("Revision", this.encode("I"))
}

// ReadICCID 获取iccid
func (this *AT) ReadICCID() (string, error) {
	if this.ICCID != "" {
		return this.ICCID, nil
	}
	command := "+ICCID"
	err := this.encodeWrite(command)
	if err != nil {
		return "", err
	}
	result, err := this.wait.Wait(command)
	if err != nil {
		return "", err
	}
	this.ICCID = result.(string)
	return this.ICCID, nil
}

// ReadIMEI 获取imei
func (this *AT) ReadIMEI() (string, error) {
	if this.IMEI != "" {
		return this.IMEI, nil
	}
	command := "+CGSN"
	err := this.encodeWrite(command)
	if err != nil {
		return "", err
	}
	result, err := this.wait.Wait(command)
	if err != nil {
		return "", err
	}
	this.IMEI = result.(string)
	return this.IMEI, nil
}

// ReadIPv4 获取ipv4信息
func (this *AT) ReadIPv4() (string, error) {
	if this.IPv4 != "" {
		return this.IPv4, nil
	}
	command := "+CGPADDR"
	result, err := this.send2(command)
	if err != nil {
		return "", err
	}
	if ls := strings.SplitN(result, ",", 2); len(ls) == 2 {
		this.IPv4 = ls[1]
		return ls[1], nil
	}
	return "", errors.New("未知错误: " + result)
}

// CSQ 网络信号 0-31越大越好
// 信号（rssi）在10～31之间均为有效值，如当地信号强的话一般不会小于20。
// 误码率直接影响信号的质量，CDMA网络信号强度99表示误码率最小；
// GPRS网络信号强度0表示误码率最小，GPRS接收信号的质量，一般要求误码<0.2%（0对应<0.2%）。大于这个值数据传输不稳定甚至不通。
func (this *AT) CSQ() (int8, int8, error) {
	command := "+CSQ"
	err := this.encodeWrite(command)
	if err != nil {
		return 0, 0, err
	}
	result, err := this.wait.Wait(command)
	if err != nil {
		return 0, 0, err
	}
	if ls := strings.Split(result.(string), ","); len(ls) == 2 {
		return conv.Int8(ls[0]), conv.Int8(ls[1]), nil
	}
	return 0, 0, errors.New("无效CSQ数据")
}

// ReadRSSI 信号（rssi）在10～31之间均为有效值，如当地信号强的话一般不会小于20。
func (this *AT) ReadRSSI() (int8, error) {
	rssi, _, err := this.CSQ()
	return rssi, err
}

// COPS 是否成功连接基站
func (this *AT) COPS() (bool, string, error) {
	command := "+COPS?"
	result, err := this.send(command, this.encode(command))
	if err != nil {
		return false, "", err
	}
	if ls := strings.Split(result, ","); len(ls) > 2 {
		return ls[1] == "1", ls[2], nil
	}
	return false, "", errors.New("未知错误: " + result)
}

// CGREG 是否成功注册网络
func (this *AT) CGREG() (bool, error) {
	command := "+CGREG?"
	result, err := this.send(command, this.encode(command))
	if err != nil {
		return false, err
	}
	if ls := strings.Split(result, ","); len(ls) > 2 {
		return ls[1] == "1", nil
	}
	return false, errors.New("未知错误: " + result)
}

// CPIN 检查SIM卡状态
func (this *AT) CPIN() (bool, error) {
	command := "+CPIN?"
	result, err := this.send(command, this.encode(command))
	if err != nil {
		return false, err
	}
	return result == "READY", nil
}

// GetTime 获取网络时间
func (this *AT) GetTime() (time.Time, error) {
	command := "+QLTS"
	result, err := this.send(command, this.encode(command+"=1")) //=0不准
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse("2006/01/02,15:04:05", result)
}
