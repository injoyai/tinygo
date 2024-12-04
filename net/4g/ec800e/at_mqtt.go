package ec800e

import (
	"errors"
	"fmt"
	"github.com/injoyai/tinygo/conv"
	"github.com/injoyai/tinygo/safe"
	"io"
	"strings"
	"time"
)

func newMQTT(at *AT, clientIndex int8) *MQTT {
	return &MQTT{
		AT:          at,
		clientIndex: clientIndex,
		subscript:   make(map[string]*MQTTSubscript),
		Closer: safe.NewCloserErr(errors.New("未初始化")).SetCloseFunc(func(err error) error {
			return at.MQTTClose(clientIndex)
		}),
	}
}

func (this *AT) MQTTDial(num int8, cfg *MQTTConfig) (err error) {

	//去除多余前缀
	address, _ := strings.CutPrefix(cfg.Address, "mqtts://")
	address, _ = strings.CutPrefix(address, "mqtt://")
	address, _ = strings.CutPrefix(address, "tcp://")
	ls := strings.SplitN(address, ":", 2)
	if len(ls) != 2 {
		return errors.New("无效MQTT地址: " + address)
	}
	//hostName 服务器地址/IP,最大长度100字节
	//port 服务器端口,范围: 1~65535
	hostName, port := ls[0], ls[1]
	command := fmt.Sprintf(`+QMTOPEN=%d,"%s",%s`, num, hostName, port)
	_, err = this.send2(command)
	if err != nil {
		return err
	}

	//等待打开异步回调
	openResult, err := this.wait.Wait("+QMTOPEN: " + conv.String(num))
	if err != nil {
		return err
	}
	if err = MQTTOpenResult(openResult.(string)).Err(); err != nil {
		return err
	}

	defer func() {
		if err != nil {
			this.MQTTClose(num)
		}
	}()

	//设置版本,3(3.1)(clientID最长只能23位), 4(3.1.1)
	command = fmt.Sprintf(`+QMTCFG="version",%d,4`, num)
	_, err = this.send2(command)
	if err != nil {
		return err
	}

	//设置清除事务
	command = fmt.Sprintf(`+QMTCFG="session",%d,%d`, num, conv.Int(cfg.CleanSession))
	_, err = this.send2(command)
	if err != nil {
		return err
	}

	//建立协议层会话
	if len(cfg.Password) > 0 {
		command = fmt.Sprintf(`+QMTCONN=%d,"%s","%s","%s"`, num, cfg.ClientID, cfg.Username, cfg.Password)
	} else if len(cfg.Username) == 0 {
		command = fmt.Sprintf(`+QMTCONN=%d,"%s"`, num, cfg.ClientID)
	} else {
		command = fmt.Sprintf(`+QMTCONN=%d,"%s","%s"`, num, cfg.ClientID, cfg.Username)
	}
	_, err = this.send2(command)
	if err != nil {
		return err
	}

	//等待连接异步回调
	result, err := this.wait.Wait("+QMTCONN: " + conv.String(num))
	if err != nil {
		return err
	}
	if err := MQTTConnectResult(result.(string)).Err(); err != nil {
		return err
	}

	return nil
}

func (this *AT) MQTTClose(clientIndex int8) error {
	if clientIndex < 0 || clientIndex > 5 {
		return errors.New("无效MQTT客户端索引: " + conv.String(clientIndex))
	}

	//断开有个过程,需要异步执行,关闭速度较慢

	////先断开会话层
	//command := "+QMTDISC=" + conv.String(clientIndex)
	//_, err := this.send2(command)
	//if err != nil {
	//	return err
	//}

	//再断开网络层
	command := "+QMTCLOSE=" + conv.String(clientIndex)
	_, err := this.send2(command)
	if err != nil {
		return err
	}

	result, err := this.wait.Wait("+QMTCLOSE: " + conv.String(clientIndex))
	if err != nil {
		return err
	}
	if err = MQTTCloseResult(result.(string)).Err(); err != nil {
		return err
	}

	return nil
}

// MQTTPublish 发布主题,分为2步骤
// 1. 告诉模块要发送的参数和数据长度
// 2. 响应">"并阻塞其他数据,写入数据后恢复正常
func (this *AT) MQTTPublish(clientIndex int8, topic string, qos uint8, retained bool, msg []byte) error {
	if clientIndex < 0 || clientIndex > 5 {
		return errors.New("无效MQTT客户端索引: " + conv.String(clientIndex))
	}

	{ //1. 告诉模块要发送的参数和数据长度,格式如下
		//+QMTPUBEX: (0-5),<msgid>,(0-2),(0,1),"topic","length"
		command := fmt.Sprintf(`+QMTPUBEX=%d,0,%d,%d,"%s",%d`,
			clientIndex, qos, conv.SelectUint8(retained, 1, 0), topic, len(msg))
		_, err := this.send2(command)
		if err != nil {
			return err
		}
	}

	{ //2. 写入数据,串口会一直阻塞等待写入数据
		_, err := this.send(string(msg), msg)
		if err != nil {
			return err
		}
		//等待执行回调
		_, err = this.wait.Wait("+QMTPUBEX: " + conv.String(clientIndex))
		if err != nil {
			return err
		}
	}

	return nil
}

// MQTTSubscribe 订阅主题,无法设置msgid,服务端可能会断开连接,不知道是不是部分模块的问题
func (this *AT) MQTTSubscribe(clientIndex int8, topic string, qos uint8) error {
	if clientIndex < 0 || clientIndex > 5 {
		return errors.New("无效MQTT客户端索引: " + conv.String(clientIndex))
	}
	m := this.MQTT[clientIndex]
	if m == nil || m.Closed() {
		return errors.New("客户端已经关闭")
	}
	//todo 这里的msgid范围是1-65535,但是设置了值后模块会重启,所以设置了0,可能服务端会校验
	command := fmt.Sprintf(`+QMTSUB=%d,0,"%s",%d`, clientIndex, topic, qos)
	_, err := this.send2(command)
	if err != nil {
		return err
	}

	//等待异步回调结果
	result, err := this.wait.Wait("+QMTSUB: " + conv.String(clientIndex))
	if err != nil {
		return err
	}
	if result.(string) != "0" {
		return errors.New("订阅失败: " + result.(string))
	}

	return nil
}

type MQTT struct {
	AT          *AT
	clientIndex int8
	subscript   map[string]*MQTTSubscript //订阅信息
	*safe.Closer
}

func (this *MQTT) Dial(cfg *MQTTConfig) error {
	if !this.Closed() {
		return nil
	}
	cfg.init()
	//可能还有链接,和程序不在同个地方,通过串口相连,忽略错误
	this.AT.MQTTClose(this.clientIndex)
	if err := this.AT.MQTTDial(this.clientIndex, cfg); err != nil {
		return err
	}
	//重置错误
	this.Closer.Reset()
	return nil
}

// Publish 发布消息
func (this *MQTT) Publish(topic string, qos uint8, retained bool, msg []byte) error {
	return this.AT.MQTTPublish(this.clientIndex, topic, qos, retained, msg)
}

// Subscribe 订阅主题
func (this *MQTT) Subscribe(topic string, qos uint8, h func(topic string, msg []byte)) error {
	err := this.AT.MQTTSubscribe(this.clientIndex, topic, qos)
	if err != nil {
		return err
	}
	this.subscript[topic] = &MQTTSubscript{
		Topic:   topic,
		Qos:     qos,
		Handler: h,
	}
	return nil
}

// UnSubscribe 取消订阅,未实现(内存中实现)
func (this *MQTT) UnSubscribe(topic string) error {
	delete(this.subscript, topic)
	return nil
}

// setSubscript 设置订阅到的数据到MQTTTopic中
func (this *MQTT) setSubscript(topic string, msg []byte) {
	for k, v := range this.subscript {
		if k == topic && v != nil {
			v.Handler(topic, msg)
		}
	}
}

type MQTTSubscript struct {
	Topic   string
	Qos     uint8
	Handler func(topic string, msg []byte)
}

/*



 */

type MQTTOpenResult string

func (this MQTTOpenResult) Err() error {
	if this == "0" {
		return nil
	}
	return this
}

func (this MQTTOpenResult) Error() string {
	switch this {
	case "-1":
		return "打开网络失败"
	case "0":
		return "成功"
	case "1":
		return "参数错误"
	case "2":
		//表示MQTT已经打开
		return "MQTT被占用"
	case "3":
		return "激活PDP失败"
	case "4":
		return "域名解析失败"
	case "5":
		return "网络断开导致错误"
	default:
		return "未知错误"
	}
}

// MQTTCloseResult 关闭MQTT实例响应 成功/失败
type MQTTCloseResult = BaseResult

// MQTTDisconnectResult MQTT断开连接响应,成功/失败
type MQTTDisconnectResult = BaseResult

type MQTTConnectState string

func (this MQTTConnectState) Error() string {
	switch this {
	case "1":
		return "MQTT初始化"
	case "2":
		return "MQTT正在连接"
	case "3":
		return "MQTT已经连接成功"
	case "4":
		return "MQTT正在断开连接"
	default:
		return "未知状态"
	}
}

type MQTTConnectResult string

func (this MQTTConnectResult) Err() error {
	if this == "0" {
		return nil
	}
	return this
}

func (this MQTTConnectResult) Error() string {
	switch this {
	case "0":
		return "成功"
	case "1":
		return "不接收的协议版本"
	case "2":
		return "标识符被拒绝"
	case "3":
		return "服务器不可用"
	case "4":
		return "错误的用户名或密码"
	case "5":
		return "未授权"
	default:
		return "未知错误"
	}
}

type BaseResult string

func (this BaseResult) Err() error {
	if this == "0" {
		return nil
	}
	return errors.New(string(this))
}

type MQTTURCErrCode string

func (this MQTTURCErrCode) Err() error {
	return this
}

func (this MQTTURCErrCode) Error() string {
	switch this {
	case "1":
		return "连接被服务器断开或者重置"
	case "2":
		return "发送PINGREQ包超时或者失败"
	case "3":
		return "发送 CONNECT 包超时或者失败"
	case "4":
		return "接收 CONNACK 包超时或者失败"
	case "5":
		//这个属于正常关闭
		return io.EOF.Error()
		return "客户端向服务器发送 DISCONNECT 包，但是服务器主动断开 MQTT 连接"
	case "6":
		return "因为发送数据包总是失败，客户端主动断开MQTT 连接"
	case "7":
		return "链路不工作或者服务器不可用"
	case "8":
		return "客户端主动断开 MQTT 连接"
	default:
		return "未知错误"
	}
}

func DefaultMQTTConfig() *MQTTConfig {
	return &MQTTConfig{
		Address:        "127.0.0.1:1883",
		ClientID:       "",
		Username:       "",
		Password:       "",
		CleanSession:   true,
		Version:        MQTTVersion_3_1_1,
		Keepalive:      0,
		ConnectTimeout: 0,
	}
}

type MQTTConfig struct {
	Address        string
	ClientID       string
	Username       string
	Password       string
	CleanSession   bool          //清除事务
	Version        MQTTVersion   //版本,推荐3.1.1
	Keepalive      time.Duration //
	ConnectTimeout time.Duration //连接超时时间
}

func (this *MQTTConfig) init() {
	if this.Version != 3 && this.Version != 4 {
		this.Version = MQTTVersion_3_1_1
	}
}

func (this *MQTTConfig) SetAddress(addr string) *MQTTConfig {
	this.Address = addr
	return this
}

func (this *MQTTConfig) SetClientID(clientID string) *MQTTConfig {
	this.ClientID = clientID
	return this
}

func (this *MQTTConfig) SetUsername(username string) *MQTTConfig {
	this.Username = username
	return this
}

func (this *MQTTConfig) SetPassword(password string) *MQTTConfig {
	this.Password = password
	return this
}

func (this *MQTTConfig) SetCleanSession(cleanSession bool) *MQTTConfig {
	this.CleanSession = cleanSession
	return this
}

func (this *MQTTConfig) SetVersion(version MQTTVersion) *MQTTConfig {
	this.Version = version
	return this
}

func (this *MQTTConfig) SetKeepalive(keepalive time.Duration) *MQTTConfig {
	this.Keepalive = keepalive
	return this
}

func (this *MQTTConfig) SetConnectTimeout(timeout time.Duration) *MQTTConfig {
	this.ConnectTimeout = timeout
	return this
}

type MQTTVersion int8

const (
	MQTTVersion_3_1   MQTTVersion = 3
	MQTTVersion_3_1_1 MQTTVersion = 4
)
