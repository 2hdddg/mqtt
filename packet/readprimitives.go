package packet

import (
	"io"
)

func (r *Reader) int2() (uint16, error) {
	msb, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	lsb, err := r.ReadByte()
	if err != nil {
		return 0, err
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
			return 0, err
		}
		if mul > (0x80 * 0x80 * 0x80) {
			return 0, newProtoErr("Var int too big")
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
		return "", err
	}
	if n != len(chars) {
		return "", newProtoErr("Too few chars in string")
	}

	return string(chars), nil
}

func (r *Reader) strMaybe() (*string, error) {
	msb, err := r.ReadByte()
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	lsb, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	l := (uint16(msb) << 8) | uint16(lsb)

	chars := make([]byte, l)
	n, err := r.Read(chars)
	if err != nil {
		return nil, err
	}
	if n != len(chars) {
		return nil, newProtoErr("Too few chars in string")
	}

	s := string(chars)
	return &s, nil
}

func (r *Reader) bin() ([]byte, error) {
	l, err := r.int2()
	if err != nil {
		return []byte{}, err
	}

	b := make([]byte, l)
	n, err := r.Read(b)
	if err != nil {
		return []byte{}, err
	}
	if n != len(b) {
		return []byte{}, newProtoErr("Too few binary bytes")
	}
	return b, nil
}

func (r *Reader) strPair() (string, string, error) {
	return "", "", nil
}
