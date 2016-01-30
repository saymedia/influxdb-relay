package main

import (
	"log"
	"net"
	"time"
)

type Reader struct {
	RecvBufferQueue chan *Message
	ListenConn      *net.UDPConn
	SendBufferQueue chan *Message
}

func (r *Reader) Run() {
	for {
		msg := <-r.RecvBufferQueue
		err := msg.ReadFromUDP(r.ListenConn)
		if err != nil {
			log.Println("error reading from UDP listener: ", err)
			// Pause so that we won't hog the CPU if we get ourselves
			// into a funny state where our UDP listener just stops
			// working for some reason.
			time.Sleep(2 * time.Second)
		}
		r.SendBufferQueue <- msg
	}
}
