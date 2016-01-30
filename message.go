package main

import (
	"net"
)

type Message struct {
	// buf is the fixed-size buffer that is allocated at the creation of
	// the message. "Content" is always a subset of buf.
	buf []byte

	// Content is the subset of "buf" that represents a single packet of data
	// from a client in InfluxDB line protocol format. The data is not
	// parsed in any way beyond the UDP framing, so it is the sender's
	// responsibility to ensure that it is valid.
	Content []byte
}

func NewMessage(maxLength int) *Message {
	ret := Message{
		buf: make([]byte, maxLength),
	}
	// Content is initially empty
	ret.Content = ret.buf[0:0]
	return &ret
}

func (m *Message) ReadFromUDP(conn *net.UDPConn) error {
	len, _, err := conn.ReadFromUDP(m.buf)
	if err != nil {
		return err
	}

	m.Content = m.buf[:len]
	return nil
}

func (m *Message) Empty() {
	m.Content = m.buf[0:0]
}
