package method

import (
	"log"

	"github.com/moxianfeng/gofactory"
)

type DumpASCII struct {
	MethodBase
}

func (this *DumpASCII) Process(in []byte) ([]byte, error) {
	s := ""
	for i := 0; i < len(in); i++ {
		s += string(in[i])
	}
	log.Print(s)
	return in, nil
}

func init() {
	gofactory.Default.Register("dumpascii", &DumpASCII{})
}
