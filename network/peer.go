package network

import (
	"fmt"
	"io"
	"sync"

	"github.com/moxianfeng/gofactory"
)

var (
	l           sync.Mutex
	freeServers map[string][]PeerInterface
)

type PeerInterface interface {
	Notify([]byte) error
	SetPeer(PeerInterface)

	NetworkInterface
}

type Peer struct {
	NetworkBase
	arguments string
	peer      PeerInterface
	loop      bool
	lock      sync.Mutex
	notify    chan []byte
}

func (this *Peer) SetPeer(p PeerInterface) {
	this.peer = p
}
func (this *Peer) Notify(data []byte) error {

	this.lock.Lock()
	defer this.lock.Unlock()

	this.notify <- data

	return nil
}

func (this *Peer) Init(arguments string) error {
	err := this.NetworkBase.Init(arguments)
	if nil != err {
		return err
	}
	this.notify = make(chan []byte, 1024)
	return nil
}

func (this *Peer) StartServer(arguments string, ch chan<- error) {
	ch <- fmt.Errorf("peer not support be server")
	return
}

func (this *Peer) Close() {
	this.loop = false
	this.notify <- []byte{}

	if nil != this.peer {
		p := this.peer
		this.peer = nil
		p.Dispatch("Error", io.EOF)
	}
}

func (this *Peer) Send(data []byte) error {
	if nil == this.peer {
		return fmt.Errorf("No peer")
	}

	return this.peer.Notify(data)
}

func (this *Peer) StartLoop() {
	this.loop = true
	// conn routine
	go func() {
		for this.loop {
			buf := <-this.notify
			// may be notify me stop loop
			if len(buf) == 0 {
				continue
			}

			newbuf, err := this.Prepare("Read", buf)
			if nil != err {
				this.Dispatch("Error", err)
			} else {
				this.Dispatch("Message", newbuf)
			}
		}
	}()
}

type PeerServer struct {
	Peer
}

type PeerClient struct {
	Peer
}

func (this *PeerServer) Dial(arguments string) error {
	// just register myself
	l.Lock()
	defer l.Unlock()

	freeServers[this.arguments] = append(freeServers[this.arguments], this)
	return nil
}

func (this *PeerClient) Dial(arguments string) error {
	// looking for freeServer
	l.Lock()
	defer l.Unlock()

	servers, ok := freeServers[this.arguments]
	if !ok {
		return fmt.Errorf("No peer ready")
	}
	if len(servers) == 0 {
		return fmt.Errorf("No peer ready")
	}

	server := servers[0]
	freeServers[this.arguments] = servers[1:]

	this.SetPeer(server)
	server.SetPeer(this)

	return nil
}

func init() {
	freeServers = make(map[string][]PeerInterface)

	gofactory.Default.Register("peerserver", &PeerServer{})
	gofactory.Default.Register("peerclient", &PeerClient{})
}
