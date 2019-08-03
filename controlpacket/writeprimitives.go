package controlpacket

func flagsToBitsU8(flags []bool) uint8 {
	bit := uint8(1)
	val := uint8(0)
	for i := len(flags) - 1; i >= 0; i-- {
		if flags[i] {
			val += bit
		}
		bit = bit << 1
	}
	return val
}

func toVarInt(val uint32) []byte {
	vals := make([]byte, 0)
	for {
		enc := uint8(val % 0x80)
		val = val / 0x80
		if val > 0 {
			enc = enc | 0x80
		}
		vals = append(vals, enc)
		if val == 0 {
			break
		}
	}
	return vals
}

func strToBytes(s string) []byte {
	l := len(s)
	b := make([]byte, l+2)
	b[0] = uint8((l >> 8) & 0xff)
	b[1] = uint8(l & 0xff)
	for i := 0; i < l; i++ {
		b[i+2] = s[i]
	}
	return b
}

func toInt2(v uint16) []byte {
	return []byte{
		uint8(v >> 8),
		uint8(v),
	}
}
