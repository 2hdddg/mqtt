package packet

func (r *Reader) int2() (uint16, error) {
	msb, err1 := r.ReadByte()
	lsb, err2 := r.ReadByte()

	if err1 != nil || err2 != nil {
		return 0, &Error{
			c: "int 2", m: "buf too small"}
	}

	return (uint16(msb) << 8) | uint16(lsb), nil
}

func (r *Reader) int4() (uint32, error) {
	return 0, nil
}

func (r *Reader) varInt() (uint32, error) {
	var mul uint32 = 1
	var val uint32 = 0

	for {
		enc, err := r.ReadByte()
		if err != nil {
			return 0, &Error{
				c: "var int", m: "buf too small"}
		}
		if mul > (0x80 * 0x80 * 0x80) {
			return 0, &Error{
				c: "var int", m: "mul overflow"}
		}

		val += uint32((enc & 0x7f)) * mul
		mul *= 0x80

		if enc <= 0x7f {
			break
		}
	}

	return val, nil
}

func (r *Reader) str() (string, error) {
	l, err := r.int2()
	if err != nil {
		return "", err
	}

	chars := make([]byte, l)
	n, err := r.Read(chars)
	if err != nil {
		return "", &Error{
			c: "string", m: "read fail", err: err}
	}
	if n != len(chars) {
		return "", &Error{
			c: "string", m: "too few bytes"}
	}

	return string(chars), nil
}

func (r *Reader) bin() ([]byte, error) {
	l, err := r.int2()
	if err != nil {
		return []byte{}, err
	}

	b := make([]byte, l)
	n, err := r.Read(b)
	if err != nil {
		return []byte{}, &Error{
			c: "binary", m: "read fail", err: err}
	}
	if n != len(b) {
		return []byte{}, &Error{
			c: "binary", m: "too few bytes"}
	}
	return b, nil
}

func (r *Reader) strPair() (string, string, error) {
	return "", "", nil
}
