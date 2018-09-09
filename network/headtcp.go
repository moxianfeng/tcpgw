package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/moxianfeng/gofactory"
)

type HeadTCP struct {
	NetworkBase
	arguments string

	conn net.Conn
	loop bool

	buffer []byte
	length int

	writeLock sync.Mutex
}

func (this *HeadTCP) StartServer(arguments string, ch chan<- error) {
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
				newconn := &HeadTCP{conn: conn}
				newconn.Init("")
				this.Dispatch("Connect", newconn)
			}
		}
	}()
}

func (this *HeadTCP) Close() {
	this.loop = false
	this.conn.Close()
}

const (
	RESERVED int = 0x193fef90
	HEADSIZE     = 8
)

func (this *HeadTCP) genHead(length int) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint32(length))
	binary.Write(buf, binary.BigEndian, uint32(RESERVED))
	return buf.Bytes()
}

func (this *HeadTCP) parseHead(buf []byte) int {
	if len(buf) < HEADSIZE {
		return 0
	}
	b := bytes.NewBuffer(buf)
	var l, r uint32
	binary.Read(b, binary.BigEndian, &l)
	binary.Read(b, binary.BigEndian, &r)
	if r != uint32(RESERVED) {
		return (int)(0xFFFFFFFF)
	}
	if l > uint32(2*1024*1024) {
		return (int)(0xFFFFFFFF)
	}
	return int(l)
}

func (this *HeadTCP) Send(data []byte) error {
	newbuf, err := this.Prepare("Write", data)
	if nil != err {
		return err
	}

	// add head
	head := this.genHead(int(len(newbuf)))

	this.writeLock.Lock()
	defer this.writeLock.Unlock()

	n, err := this.conn.Write(head)
	if nil != err {
		return err
	}
	if n != len(head) {
		return fmt.Errorf("Send error")
	}

	n, err = this.conn.Write(newbuf)
	if nil != err {
		return err
	}

	if n != len(newbuf) {
		return fmt.Errorf("Send error")
	}

	return nil
}

func (this *HeadTCP) Dial(arguments string) error {
	this.arguments = arguments
	conn, err := net.Dial("tcp", arguments)
	if nil != err {
		return err
	}
	this.conn = conn

	return nil
}

func (this *HeadTCP) StartLoop() {
	this.loop = true
	if nil == this.buffer {
		this.buffer = make([]byte, 2*1024*1024)
	}
	// conn routine
	go func() {
		for this.loop {
			n, err := this.conn.Read(this.buffer[this.length:])
			if nil != err {
				this.Dispatch("Error", err)
			} else {
				this.length += n

				// process message
				for {
					pkglen := this.parseHead(this.buffer[:this.length])
					if pkglen == 0 {
						// 长度不足包头长度
						break
					}

					if pkglen == (int)(0xFFFFFFFF) {
						// 包错乱
						this.Dispatch("Error", fmt.Errorf("May be wrong order"))
						break
					}

					if this.length >= pkglen+HEADSIZE {
						newbuf, err := this.Prepare("Read", this.buffer[HEADSIZE:HEADSIZE+pkglen])
						if nil != err {
							this.Dispatch("Error", err)
							break
						} else {
							this.Dispatch("Message", newbuf)
							remain := this.length - HEADSIZE - pkglen
							if remain == 0 {
								this.length = 0
								// 没有内容了
								break
							} else {
								rebuf := make([]byte, 2*1024*1024)
								copy(rebuf, this.buffer[HEADSIZE+pkglen:this.length])
								this.length = remain
								this.buffer = rebuf
								// 继续处理消息
							}
						}
					} else {
						// 长度不足
						break
					}
				}

			}
		}
	}()
}

func init() {
	gofactory.Default.Register("head", &HeadTCP{})
}
