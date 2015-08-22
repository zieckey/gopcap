package main

import (
	"fmt"
	"net"
	"time"
)

const (
	CloseTunnel = 1
)

type Tunnel struct {
	conn []net.Conn

	dch chan []byte // data channel to transfer receiving data
	ach chan int    // admin channel
	timeout time.Duration
}

func NewTunnel(count int, protocol string, addr string) *Tunnel {
	t := new(Tunnel)
	t.dch = make(chan []byte, 5)
	t.ach = make(chan int)
	t.timeout = time.Duration(100) * time.Second

	for c := 0; c < count; c++ {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			fmt.Printf("[%s] connect to %s failed : %v\n", protocol, addr, err.Error())
			return nil
		}
		t.conn = append(t.conn, conn)

		go t.read(conn)
		go t.serve()
	}

	return t
}

func (t *Tunnel) read(conn net.Conn) {
	buf := make([]byte, 65536)
	for {
		conn.SetReadDeadline(time.Now().Add(t.timeout))
		n, e := conn.Read(buf)
		if n == 0 || e != nil {
			fmt.Printf("stop reading data. n=%d err=%v\n", n, e.Error())
			break
		}
	}
}

func (t *Tunnel) serve() {
	for {
		select {
		case data := <-t.dch:
			for _, c := range t.conn {
				c.Write(data)
			}
		case closing := <-t.ach:
			if closing == CloseTunnel {
				for _, c := range t.conn {
					c.Close()
				}
			}
		}

	}
}

func (t *Tunnel) Write(d []byte) {
	t.dch <- d
}

func (t *Tunnel) Close() {
	t.ach <- CloseTunnel
	t.timeout = time.Second
}
