package ipcbench

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

type UnixNotify struct {
	Uri      string // Socket address
	Protocol string
	Messages int // Number of test messages used to test latency

	active bool
	net.Conn
	sync.Mutex
}

func (u *UnixNotify) CreateNotifier(uri, protocol string, messages int) error {
	u.Lock()
	defer u.Unlock()

	if u.active {
		return fmt.Errorf("Error: Socket active ", uri)
	}

	u.Uri = uri
	u.Protocol = protocol
	u.Messages = messages

	for {
		c, err := net.Dial(u.Protocol, u.Uri)

		if err != nil {
			continue
		}
		u.Conn = c

		u.active = true
		return nil
	}
}

func (u *UnixNotify) CreateListener(uri, protocol string, messages int) error {
	u.Lock()
	defer u.Unlock()

	if u.active {
		return fmt.Errorf("Error: Socket active ", uri)
	}

	u.Uri = uri
	u.Protocol = protocol
	u.Messages = messages

	l, err := net.Listen(u.Protocol, u.Uri)
	if err != nil {
		return err
	}

	// Just handle one connection for now
	for {
		fd, err := l.Accept()
		if err != nil {
			return err
		}
		u.Conn = fd
		u.active = true
		return nil
	}

	return nil
}

func (u *UnixNotify) Close() error {
	u.Lock()
	defer u.Unlock()

	if !u.active {
		fmt.Println("ERROR: Not active")
		return nil
	}
	u.active = false
	err := u.Conn.Close()
	return err
}

func (u *UnixNotify) Notify() (minLatency, maxLatency, avgLatency time.Duration, latencies []time.Duration, err error) {
	buf := make([]byte, 32)
	var max, avg time.Duration
	min := time.Hour

	for i := 0; i < u.Messages; i++ {
		then := time.Now()
		n := binary.PutVarint(buf, then.UnixNano())

		// Send it out and wait for the reply
		// We are measuring round trip time
		// as we want to use the same reference time
		u.Write(buf[:n])
		n, err = u.Read(buf)
		if err != nil {
			return 0, 0, 0, nil, fmt.Errorf("ERROR:", err)
		}
		now := time.Now()

		reply, n := binary.Varint(buf)
		if n <= 0 {
			return 0, 0, 0, nil, fmt.Errorf("ERROR: Invalid reply")
		}
		if reply != then.UnixNano() {
			return 0, 0, 0, nil, fmt.Errorf("ERROR: Invalid reply: Token mismatch", reply, then.UnixNano())
		}

		latency := now.Sub(then)
		avg = avg + latency

		latencies = append(latencies, latency)

		if latency > max {
			max = latency
		}
		if latency < min {
			min = latency
		}
	}

	totalLat := time.Duration((avg.Nanoseconds() / int64(u.Messages)))

	return min, max, totalLat, latencies, nil
}

func (u *UnixNotify) Wait() error {
	buf := make([]byte, 1024)

	for i := 0; i < u.Messages; i++ {
		n, err := u.Read(buf)
		if err != nil {
			return err
		}
		u.Write(buf[:n])
	}
	return nil
}
