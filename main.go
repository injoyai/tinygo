package main

import (
	"github.com/injoyai/tinygo/cfg"
	"github.com/injoyai/tinygo/conv"
	"github.com/injoyai/tinygo/logs"
	"github.com/injoyai/tinygo/net/4g/ec800e"
	"github.com/injoyai/tinygo/oss/flash"
	"github.com/injoyai/tinygo/oss/led"
	"github.com/injoyai/tinygo/oss/serial"
	"github.com/injoyai/tinygo/protocol/easy"
	"github.com/injoyai/tinygo/times"
	"machine"
	"runtime"
	"time"
)

var (
	Server    = easy.NewServer()
	AT        *ec800e.AT
	CfgSerial *serial.Serial
)

func init() {
	CfgSerial = serial.New(machine.UART0, machine.GP0, machine.GP1)
	logs.SetWriter(CfgSerial)

	//初始化配置文件
	cfg.Init(flash.New(flash.PicoStart, 4))

	netSerial := serial.New(machine.UART1, machine.GP8, machine.GP9)
	netSerial.SetBaudRate(115200)
	AT = ec800e.New(netSerial, time.Second*2)

	Server.RegisterGet("chip_version", func() any { return machine.ChipVersion() })
	Server.RegisterGet("temp", func() any { return float32(machine.ReadTemperature()) / 1000 })
	Server.RegisterGet("core_current", func() any { return machine.CurrentCore() })
	Server.RegisterGet("core_num", func() any { return machine.NumCores() })
	Server.RegisterGet("cpu", func() any { return machine.CPUFrequency() })
	Server.RegisterGet("cpu_num", func() any { return runtime.NumCPU() })
	Server.RegisterGet("os", func() any { return runtime.GOOS })
	Server.RegisterGet("arch", func() any { return runtime.GOARCH })
	Server.RegisterGet("version", func() any { return runtime.Version() }) //tinygo版本
	Server.RegisterGet("root", func() any { return runtime.GOROOT() })
	Server.RegisterGet("go_num", func() any { return runtime.NumGoroutine() })
	Server.RegisterGet("time", func() any { return times.Now().Format(time.DateTime) })
	Server.RegisterGet("4g_imei", func() any {
		imei, err := AT.ReadIMEI()
		if err != nil {
			return err
		}
		return imei
	})
	Server.RegisterGet("4g_iccid", func() any {
		iccid, err := AT.ReadICCID()
		if err != nil {
			return err
		}
		return iccid
	})
	Server.RegisterGet("4g_ip", func() any {
		ip, err := AT.ReadIPv4()
		if err != nil {
			return err
		}
		return ip
	})
	Server.RegisterGet("4g_rssi", func() any {
		rssi, err := AT.ReadRSSI()
		if err != nil {
			return err
		}
		return rssi
	})
	Server.RegisterCfg("test")

}

func main() {

	l := led.New(machine.LED).GoBlink(.5, .5)
	_ = l

	go Server.Bridge(CfgSerial)

	//runUpload(AT, &MQTTConfig{
	//	Topic:  "TOPIC",
	//	Qos:    0,
	//	Retain: false,
	//	Cycle:  time.Second * 10,
	//	MQTTConfig: &ec800e.MQTTConfig{
	//		Address:  "39.107.120.124:11883",
	//		ClientID: "PICO_TEST",
	//	},
	//})

	select {}

}

func runUpload(at *ec800e.AT, cfg *MQTTConfig) {

	var client = at.GetMQTT(4)
	var msgid = 0

	for {

		//建立连接
		if client != nil && client.Closed() {
			err := client.Dial(cfg.MQTTConfig)
			if err != nil {
				logs.Err(err)
			}
		}

		//推送消息
		if client != nil && !client.Closed() {
			msg := conv.String(map[string]any{
				"data": map[string]any{
					"time":  time.Now().String(),
					"msgID": msgid,
				},
				"ts": time.Now().UnixMilli(),
			})
			err := client.Publish(cfg.Topic, cfg.Qos, cfg.Retain, []byte(msg))
			if err != nil {
				logs.Err(err)
			}
		}

		<-time.After(cfg.Cycle)

	}

}

type MQTTConfig struct {
	Topic  string
	Qos    uint8
	Retain bool
	Cycle  time.Duration
	*ec800e.MQTTConfig
}
