package main

import (
	"bytes"
	"log"
	"net/http"
	"time"
)

type Writer struct {
	SendBufferQueue chan *Message
	TargetURL       string
	RecvBufferQueue chan *Message
}

func (w *Writer) Run() {
	transport := http.Transport{}
	client := http.Client{
		Transport: &transport,
		Timeout:   2 * time.Second,
	}

	// Only stack allocation is allowed inside this loop.
	// When calling functions that allocate and return pointers, ensure
	// that golang's escape analysis can tell that the memory can
	// actually be allocated on the stack.
	// (Use -gcflags '-m -l' to prove this.)
	for {
		msg := <-w.SendBufferQueue

		// We'll keep trying to send this message until we succeed
		// or get tired of trying
		tries := 0
		for {
			tries += 1

			// we tried, really hard, but let's be serious...
			if tries == *attemptLimit {
				log.Println("gave up writing to backend after", tries, "attempts")
				break
			}

			contentReader := bytes.NewReader(msg.Content)

			resp, err := client.Post(
				w.TargetURL, "application/octet-stream", contentReader,
			)
			if err != nil {
				log.Println("error writing to backend:", err)
				time.Sleep(2 * time.Second)
				continue
			}

			// InfluxDB docs say that only 204 exactly is truly successful,
			// and in fact 200 OK is not successful. Strange, but okay...
			if resp.StatusCode != 204 {
				log.Println("backend write returned", resp.Status)
				if resp.StatusCode == 400 {
					// invalid request; no amount of retries will
					// reform malformed data
					break
				}

				time.Sleep(2 * time.Second)
				continue
			}

			// If we fell out here then we've succeeded.
			break
		}

		// Put the buffer back in the recieve queue so it can be re-used
		// for a future message.
		msg.Empty()
		w.RecvBufferQueue <- msg

	}
}
