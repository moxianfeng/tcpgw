package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"

	m "github.com/moxianfeng/tcpgw/method"
	n "github.com/moxianfeng/tcpgw/network"

	"github.com/moxianfeng/gofactory"
)

var v = gofactory.Default

type FileConfig struct {
	Methods string         `json:"method"`
	Servers []ServerConfig `json:"servers"`
}

type ServerConfig struct {
	Name     string        `json:"name"`
	Frontend NetworkConfig `json:"frontend"`
	Backend  NetworkConfig `json:"backend"`
}

type NetworkConfig struct {
	Network     NetworkType `json:"net"`
	AfterRead   Action      `json:"afterread"`
	BeforeWrite Action      `json:"beforewrite"`
}

type NetworkType struct {
	Type string `json:"type"`
	Addr string `json:"addr"`
}

type Action struct {
	Method    string `json:"method"`
	Arguments string `json:"arguments"`
}

func _singlePrepare(net n.NetworkInterface, action Action, event string) error {
	method, err := gofactory.Default.GetInterface(action.Method)
	if nil != err {
		return err
	}
	methodInstance, err := m.NewMethod(method)
	if nil != err {
		return err
	}

	methodInstance.SetArguments(action.Arguments)
	net.OnPrepare(event, func(event string, data []byte) ([]byte, error) {
		return methodInstance.Process(data)
	})
	return nil
}

func setPrepare(net n.NetworkInterface, config NetworkConfig) error {

	err := _singlePrepare(net, config.AfterRead, "Read")
	if nil != err {
		return err
	}

	err = _singlePrepare(net, config.BeforeWrite, "Write")
	if nil != err {
		return err
	}
	return nil
}

func startServer(server ServerConfig, ch chan<- error) {
	front, err := gofactory.Default.GetInterface(server.Frontend.Network.Type)
	if nil != err {
		ch <- err
		return
	}

	frontserver, err := n.NewNetwork(front, server.Frontend.Network.Addr)
	if nil != err {
		ch <- err
		return
	}

	frontserver.On("Connect", func(event string, data interface{}) {
		frontend := data.(n.NetworkInterface)
		err = setPrepare(frontend, server.Frontend)
		if nil != err {
			log.Printf("setPrepare for frontend failed, %v", err)
			frontend.Close()
			return
		}

		backclient, err := gofactory.Default.GetInterface(server.Backend.Network.Type)
		if nil != err {
			log.Printf("Can't get backend `%s`, %v", server.Backend.Network.Type, err)
			frontend.Close()
			return
		}
		backend, err := n.NewNetwork(backclient, server.Backend.Network.Addr)
		if nil != err {
			log.Printf("Can't new backend `%s`, %v", server.Backend.Network.Type, err)
			frontend.Close()
			return
		}

		err = setPrepare(backend, server.Backend)
		if nil != err {
			log.Printf("setPrepare for backend failed, %v", err)
			frontend.Close()
			return
		}

		err = backend.Dial(server.Backend.Network.Addr)
		if nil != err {
			log.Printf("Dial backend `%s:%s`, %v", server.Backend.Network.Type, server.Backend.Network.Addr, err)
			frontend.Close()
			return
		}

		// Close
		OnError := func(event string, data interface{}) {
			// log.Print(data)
			frontend.Close()
			backend.Close()
		}

		OnMessage := func(peer n.NetworkInterface) func(event string, data interface{}) {
			return func(event string, data interface{}) {
				err := peer.Send(data.([]byte))
				if nil != err {
					log.Printf("Send to peer error, %v", err)
					frontend.Close()
					backend.Close()
				}
			}
		}

		frontend.On("Error", OnError)
		backend.On("Error", OnError)
		frontend.On("Message", OnMessage(backend))
		backend.On("Message", OnMessage(frontend))

		backend.StartLoop()
		frontend.StartLoop()
	})

	frontserver.StartServer(server.Frontend.Network.Addr, ch)
}

func main() {
	config := flag.String("config", "/etc/tcpgw.json", "The config file")
	flag.Parse()

	f, err := os.Open(*config)
	if nil != err {
		log.Fatal(err)
	}
	defer f.Close()

	content, err := ioutil.ReadAll(f)
	if nil != err {
		log.Fatal(err)
	}

	var fc FileConfig

	err = json.Unmarshal(content, &fc)
	if nil != err {
		log.Fatal(err)
	}

	serverCount := len(fc.Servers)
	waitingCount := serverCount

	ch := make(chan error, 10)
	for _, server := range fc.Servers {
		go startServer(server, ch)
	}

	for waitingCount > 0 {
		exitErr := <-ch
		log.Print(exitErr)

		waitingCount--
		if waitingCount == 0 {
			break
		}
	}
	log.Print("all servers finished")
}
