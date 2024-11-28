package serial

import "machine"

type Serial struct {
	machine.Serialer
}

// Read 实现io.Reader接口
func (this *Serial) Read(p []byte) (n int, err error) {
	for n = 0; n < len(p) || n < this.Buffered(); n++ {
		p[n], err = this.Serialer.ReadByte()
		if err != nil {
			return 0, err
		}
	}
	return
}

type Config struct {
	BaudRate uint32
	RX       machine.Pin
	TX       machine.Pin
}
