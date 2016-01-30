package main

import (
	"flag"
	"fmt"
	"net"
	"os"
)

var maxPacketSize = flag.Int(
	"max-line-length", 256, "Maximum line length for line protocol, in bytes",
)
var bufferSize = flag.Int(
	"buffer-size", 4096, "Maximum number of packets that can be buffered",
)
var targetURL = flag.String(
	"target-url", "http://127.0.0.1:8086/write?db=example", "URL where recieved data should be written",
)
var listenAddrStr = flag.String(
	"listen-addr", "127.0.0.1:4444", "Local address for the UDP listener",
)

func main() {
	flag.Parse()

	listenAddr, err := net.ResolveUDPAddr("udp", *listenAddrStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error resolving -listen-addr: %s", err)
		os.Exit(1)
	}

	listenConn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen on %s: %s", listenAddrStr, err)
		os.Exit(2)
	}
	defer listenConn.Close()

	// recvBufferQueue is a FIFO queue of "empty" Messages that are ready
	// to be used to recieve data. At startup this buffer is filled with
	// instances, which are taken by the reader and then returned to this
	// queue by the writer once a message has been transmitted ot the backend.
	recvBufferQueue := make(chan *Message, *bufferSize)

	// sendBufferQueue is a FIFO queue of read Messages that are ready to be
	// transmitted to the backend.
	sendBufferQueue := make(chan *Message, *bufferSize)

	// Allocate all of the message objects we'll use. By allocating these
	// up front we avoid doing heap allocations after we initialize, thus
	// reducing GC pressure.
	// These all go into the recvBufferQueue where they can be read by
	// the reader goroutine when it starts up.
	for i := 0; i < *bufferSize; i++ {
		msg := NewMessage(*maxPacketSize)
		recvBufferQueue <- msg
	}

	reader := &Reader{
		RecvBufferQueue: recvBufferQueue,
		ListenConn:      listenConn,
		SendBufferQueue: sendBufferQueue,
	}
	writer := &Writer{
		SendBufferQueue:   sendBufferQueue,
		TargetURL:         *targetURL,
		RecvBufferQueue:   recvBufferQueue,
	}

	go reader.Run()
	writer.Run()
}
