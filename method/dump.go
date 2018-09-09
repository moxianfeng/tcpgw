package method

import (
	"log"

	"github.com/moxianfeng/gofactory"
)

type Dump struct {
	MethodBase
}

func (this *Dump) Process(in []byte) ([]byte, error) {
	log.Print(in)
	return in, nil
}

func init() {
	gofactory.Default.Register("dump", &Dump{})
}
