package modbus

func NewServer() *Server {
	s := &Server{}
	s.SetHandler(1, s.handler1)
	s.SetHandler(2, s.handler2)
	s.SetHandler(3, s.handler3)
	s.SetHandler(4, s.handler4)
	s.SetHandler(5, s.handler5)
	s.SetHandler(6, s.handler6)
	s.SetHandler(15, s.handler15)
	s.SetHandler(16, s.handler16)
	return s
}

type Server struct {
	Coils            Coils       //线圈 0x01(0x读),0x05(写1个),0x0f(写多个)
	DiscreteInputs   Coils       //离散输入(只读的线圈) 0x02(1x读)
	InputRegisters   Register    //输入寄存器 0x04(3x读)
	HoldingRegisters Register    //保持寄存器 0x03(4x读) 0x06(写1个) 0x10(写多个)
	Handler          [43]Handler //处理函数,下标对应功能码
}

func (this *Server) Deal(bs []byte) ([]byte, error) {
	f, err := DecodeRTU(bs)
	if err != nil {
		return nil, err
	}
	switch f.GetControl() {
	case 1, 2, 3, 4, 5, 6, 15, 16:
		handler := this.Handler[f.GetControl()]
		if handler != nil {
			result, control := handler(f)
			if control != Success {
				f.SetControl(control)
			}
			f.SetData(result)
		}
	default:
		f.SetControl(IllegalFunction)
	}
	return f.Bytes(), nil
}

// SetCoils 设置线圈接口
func (this *Server) SetCoils(register uint16, wrc ReadWriteCoils) *Server {
	this.Coils[register] = wrc
	return this
}

// SetDiscreteInputs 设置离散输入接口
func (this *Server) SetDiscreteInputs(register uint16, wrc ReadWriteCoils) *Server {
	this.DiscreteInputs[register] = wrc
	return this
}

// SetInputRegisters 设置数据寄存器接口
func (this *Server) SetInputRegisters(register uint16, wrc ReadWriteRegister) *Server {
	this.InputRegisters[register] = wrc
	return this
}

// SetHoldingRegisters 设置保持寄存器接口
func (this *Server) SetHoldingRegisters(register uint16, wrc ReadWriteRegister) *Server {
	this.HoldingRegisters[register] = wrc
	return this
}

// SetHandler 设置功能码对应函数
func (this *Server) SetHandler(code int, handler Handler) {
	if code > 0 && code < len(this.Handler) {
		this.Handler[code] = handler
	}
}
