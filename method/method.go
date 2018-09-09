package method

import (
	"fmt"
	"reflect"
)

type MethodInterface interface {
	Process([]byte) ([]byte, error)
	SetArguments(arguments string)
}

type MethodBase struct {
	Arguments string
}

var methodInterfaceType = reflect.TypeOf(new(MethodInterface)).Elem()

func NewMethod(sample interface{}) (MethodInterface, error) {
	sampleType := reflect.TypeOf(sample)
	if sampleType.Kind() != reflect.Ptr {
		sampleType = reflect.PtrTo(sampleType)
	}
	if sampleType.Implements(methodInterfaceType) {
		return reflect.New(sampleType.Elem()).Interface().(MethodInterface), nil
	} else {
		return nil, fmt.Errorf("Type not match")
	}
}

func (this *MethodBase) SetArguments(arguments string) {
	this.Arguments = arguments
}
