package packet

type ProtocolError struct {
	m string
}

func newProtoErr(m string) *ProtocolError {
	return &ProtocolError{m: m}
}

func (e *ProtocolError) Error() string {
	return e.m
}


