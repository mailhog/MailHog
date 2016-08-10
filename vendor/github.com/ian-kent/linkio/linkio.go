// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package linkio provides an io.Reader and io.Writer that
// simulate a network connection of a certain speed.
package linkio

import (
	"io"
	"time"
)

// Throughput represents the link speed as an int64 bits per second
// count. The representation limits the largest representable throughput
// to approximately 9223 petabits per second.
type Throughput int64

// Common throughputs.
//
// To count the number of units in a Duration, divide:
//	kilobit := linkio.KilobitPerSecond
//	fmt.Print(int64(kilobit/linkio.BitPerSecond)) // prints 1024
//
// To convert an integer number of units to a Throughput, multiply:
//	megabits := 10
//	fmt.Print(linkio.Throughput(megabits)*time.BitPerSecond) // prints 10s
//
const (
	BitPerSecond      Throughput = 1
	BytePerSecond                = 8 * BitPerSecond
	KilobitPerSecond             = 1024 * BitPerSecond
	KilobytePerSecond            = 1024 * BytePerSecond
	MegabitPerSecond             = 1024 * KilobitPerSecond
	MegabytePerSecond            = 1024 * KilobytePerSecond
	GigabitPerSecond             = 1024 * MegabitPerSecond
	GigabytePerSecond            = 1024 * MegabytePerSecond
)

// A LinkReader wraps an io.Reader, simulating reading from a
// shared access link with a fixed maximum speed.
type LinkReader struct {
	r    io.Reader
	link *Link
}

// A LinkWriter wraps an io.Writer, simulating writer to a
// shared access link with a fixed maximum speed.
type LinkWriter struct {
	w    io.Writer
	link *Link
}

// A Link serializes requests to sleep, simulating the way data travels
// across a link which is running at a certain kbps (kilo = 1024).
// Multiple LinkReaders can share a link (simulating multiple apps
// sharing a link). The sharing behavior is approximately fair, as implemented
// by Go when scheduling reads from a contested blocking channel.
type Link struct {
	in    chan linkRequest
	out   chan linkRequest
	speed int64 // nanosec per bit
}

// A linkRequest asks the link to simulate sending that much data
// and return a true on the channel when it has accomplished the request.
type linkRequest struct {
	bytes int
	done  chan bool
}

// NewLinkReader returns a LinkReader that returns bytes from r,
// simulating that they arrived from a shared link.
func (link *Link) NewLinkReader(r io.Reader) (s *LinkReader) {
	s = &LinkReader{r: r, link: link}
	return
}

// NewLinkWriter returns a LinkWriter that writes bytes to r,
// simulating that they arrived from a shared link.
func (link *Link) NewLinkWriter(w io.Writer) (s *LinkWriter) {
	s = &LinkWriter{w: w, link: link}
	return
}

// NewLink returns a new Link running at kbps.
func NewLink(throughput Throughput) (l *Link) {
	// allow up to 100 outstanding requests
	l = &Link{in: make(chan linkRequest, 100), out: make(chan linkRequest, 100)}
	l.SetThroughput(throughput)

	// This goroutine serializes the requests. He could calculate
	// link utilization by comparing the time he sleeps waiting for
	// linkRequests to arrive and the time he spends sleeping to simulate
	// traffic flowing.

	go func() {
		for lr := range l.in {
			// bits * nanosec/bit = nano to wait
			delay := time.Duration(int64(lr.bytes*8) * l.speed)
			time.Sleep(delay)
			lr.done <- true
		}
	}()
	go func() {
		for lr := range l.out {
			// bits * nanosec/bit = nano to wait
			delay := time.Duration(int64(lr.bytes*8) * l.speed)
			time.Sleep(delay)
			lr.done <- true
		}
	}()

	return
}

// SetThroughput sets the current link throughput
func (link *Link) SetThroughput(throughput Throughput) {
	// link.speed is stored in ns/bit
	link.speed = 1e9 / int64(throughput)
}

// why isn't this in package math? hmm.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Satisfies interface io.Reader.
func (l *LinkReader) Read(buf []byte) (n int, err error) {
	// Read small chunks at a time, even if they ask for more,
	// preventing one LinkReader from saturating the simulated link.
	// 1500 is the MTU for Ethernet, i.e. a likely maximum packet
	// size.
	toRead := min(len(buf), 1500)
	n, err = l.r.Read(buf[0:toRead])
	if err != nil {
		return 0, err
	}

	// send in the request to sleep to the Link and sleep
	lr := linkRequest{bytes: n, done: make(chan bool)}
	l.link.in <- lr
	_ = <-lr.done

	return
}

// Satisfies interface io.Writer.
func (l *LinkWriter) Write(buf []byte) (n int, err error) {
	// Write small chunks at a time, even if they attempt more,
	// preventing one LinkReader from saturating the simulated link.
	// 1500 is the MTU for Ethernet, i.e. a likely maximum packet
	// size.
	toWrite := min(len(buf), 1500)
	n, err = l.w.Write(buf[0:toWrite])
	if err != nil {
		return 0, err
	}

	// send in the request to sleep to the Link and sleep
	lr := linkRequest{bytes: n, done: make(chan bool)}
	l.link.in <- lr
	_ = <-lr.done

	return
}
