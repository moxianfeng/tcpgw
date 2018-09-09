package network

import (
	"fmt"
	"reflect"
)

type NetworkCallback func(event string, data interface{})
type PrepareCallback func(event string, data []byte) ([]byte, error)

type NetworkBase struct {
	EventsMap  map[string]NetworkCallback
	PrepareMap map[string]PrepareCallback
	Address    string
}

func (this *NetworkBase) Init(addr string) error {
	this.Address = addr
	this.EventsMap = make(map[string]NetworkCallback)
	this.PrepareMap = make(map[string]PrepareCallback)
	return nil
}

func (this *NetworkBase) On(event string, callback NetworkCallback) {
	this.EventsMap[event] = callback
}

func (this *NetworkBase) OnPrepare(event string, callback PrepareCallback) {
	this.PrepareMap[event] = callback
}

func (this *NetworkBase) Prepare(event string, data []byte) ([]byte, error) {
	callback, ok := this.PrepareMap[event]
	if !ok {
		return data, nil
	}
	return callback(event, data)
}

func (this *NetworkBase) Dispatch(event string, data interface{}) int {
	cb, ok := this.EventsMap[event]
	if !ok {
		return -1
	}
	cb(event, data)
	return 0
}

type NetworkInterface interface {
	Init(addr string) error
	StartServer(arguments string, ch chan<- error)
	StartLoop()

	Dial(arguments string) error

	On(event string, callback NetworkCallback)
	OnPrepare(event string, callback PrepareCallback)

	Dispatch(event string, data interface{}) int

	Send([]byte) error
	Close()
}

var networkInterfaceType = reflect.TypeOf(new(NetworkInterface)).Elem()

func NewNetwork(sample interface{}, addr string) (NetworkInterface, error) {
	sampleType := reflect.TypeOf(sample)
	if sampleType.Kind() != reflect.Ptr {
		sampleType = reflect.PtrTo(sampleType)
	}
	if sampleType.Implements(networkInterfaceType) {
		ret := reflect.New(sampleType.Elem()).Interface().(NetworkInterface)
		err := ret.Init(addr)
		if nil != err {
			return nil, err
		}
		return ret, nil
	} else {
		return nil, fmt.Errorf("Type not match")
	}
}
