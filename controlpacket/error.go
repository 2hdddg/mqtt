package controlpacket

import "fmt"

type Error struct {
	c   string
	m   string
	err error
}

func (e *Error) Error() string {
	return fmt.Sprintf("Control packet error, %s: %s, %s", e.c, e.m, e.err)
}

