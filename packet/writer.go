package packet

import (
	"errors"
	"io"
)

type Writer struct {
	io.Writer
}

func (w *Writer) WriteAckConnection(ack *AckConnection) error {
	if ack == nil {
		return errors.New("nil")
	}
	packet := ack.toPacket()
	_, err := w.Write(packet)
	if err != nil {
		return err
	}
	return nil
}

func (w *Writer) WritePingResp(resp *PingResp) error {
	packet := resp.toPacket()
	_, err := w.Write(packet)
	return err
}
