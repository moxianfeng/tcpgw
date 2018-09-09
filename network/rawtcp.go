package network

import (
	"fmt"
	"log"
	"net"

	"github.com/moxianfeng/gofactory"
)

type RawTCP struct {
	NetworkBase
	arguments string

	conn net.Conn
	loop bool
}

func (this *RawTCP) StartServer(arguments string, ch chan<- error) {
	this.arguments = arguments

	ln, err := net.Listen("tcp", this.arguments)
	if nil != err {
		ch <- err
		return
	}

	// server routine
	go func() {
		for {
			conn, err := ln.Accept()
			if nil != err {
				this.Dispatch("Error", err)
				log.Printf("Accept return %v", err)
			} else {
				newconn := &RawTCP{conn: conn}
				newconn.Init("")
				this.Dispatch("Connect", newconn)
			}
		}
	}()
}

func (this *RawTCP) Close() {
	this.loop = false
	this.conn.Close()
}

func (this *RawTCP) Send(data []byte) error {
	newbuf, err := this.Prepare("Write", data)
	if nil != err {
		return err
	}

	n, err := this.conn.Write(newbuf)
	if nil != err {
		return err
	}

	if n != len(newbuf) {
		return fmt.Errorf("Send error")
	}
	return nil
}

func (this *RawTCP) Dial(arguments string) error {
	this.arguments = arguments
	conn, err := net.Dial("tcp", arguments)
	if nil != err {
		return err
	}
	this.conn = conn

	return nil
}

func (this *RawTCP) StartLoop() {
	this.loop = true
	// conn routine
	go func() {
		var buf []byte = make([]byte, 1024*1024)
		for this.loop {
			n, err := this.conn.Read(buf)
			if nil != err {
				this.Dispatch("Error", err)
			} else {
				newbuf, err := this.Prepare("Read", buf[:n])
				if nil != err {
					this.Dispatch("Error", err)
				} else {
					this.Dispatch("Message", newbuf)
				}
			}
		}
	}()
}

func init() {
	gofactory.Default.Register("raw", &RawTCP{})
}
