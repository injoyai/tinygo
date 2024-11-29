package easy

func Sum(bs []byte) uint8 {
	b := uint8(0)
	for i := range bs {
		b += bs[i]
	}
	return b
}
