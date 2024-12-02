package main

import (
	"github.com/injoyai/tinygo/logs"
	"github.com/injoyai/tinygo/os/led"
	"github.com/injoyai/tinygo/os/serial"
	"github.com/injoyai/tinygo/protocol/modbus"
	"io"
	"machine"
	"runtime"
	"time"
)

var (
	Server = modbus.NewServer()
)

func init() {
	Server.SetHoldingRegisters(255, modbus.ReadWriteRegister{
		Read: func() (result [2]byte, err error) {
			return [2]byte{0x01, 0x02}, nil
		},
		Write: func(data [2]byte) (err error) {
			return nil
		},
	})
}

func main() {

	l := led.New(machine.LED).GoBlink(.5, .5)

	p := serial.New(machine.UART1, machine.GP4, machine.GP5)

	_ = l
	_ = p

	logs.SetWriter(p)

	s := serial.NewServer(p, 128, func(s *serial.Server) {
		s.OnDealMessage = func(bs []byte) {
			logs.Read(string(bs))
		}
		s.GoTimerWriter(time.Second*5, func(w io.Writer) {
			w.Write([]byte("111"))
			logs.Info(runtime.Version())
			logs.Debug(machine.CPUFrequency())
		})
	})

	go s.Run()

	for {
		<-time.After(time.Second * 5)
		//logs.Debug("debug")
		l.Insert(led.BlinkFast)
		p.Write([]byte("hello"))
	}

	select {}

}
